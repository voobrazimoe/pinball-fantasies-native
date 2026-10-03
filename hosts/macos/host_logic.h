#ifndef PF_MAC_HOST_LOGIC_H
#define PF_MAC_HOST_LOGIC_H
#include <stdbool.h>
#include <stdint.h>
#include <stddef.h>
#include <stdatomic.h>
#include "../../cmd/pfengine/abi.h"
/* Native keycodes live only here. Output is exclusively ABI logical input. */
typedef enum { PF_EVENT_ACTION, PF_EVENT_KEY, PF_EVENT_RELEASE, PF_EVENT_FIRE,
               PF_EVENT_DELTA, PF_EVENT_FULLSCREEN } PFHostEvent;
typedef void (*PFEmit)(void *, PFHostEvent, int32_t, int32_t);
typedef struct {
    bool down[128], held[4], fire, focused, fullscreenPending;
    PFEmit emit;
    void *context;
} PFInput;
void pf_input_init(PFInput *, PFEmit, void *);
void pf_input_key(PFInput *, uint16_t, bool down, bool repeat, bool command, bool option);
void pf_input_modifiers(PFInput *, const bool sides[6]);
void pf_input_focus(PFInput *, bool);
void pf_input_button(PFInput *, bool down, bool mouseActive);
void pf_input_motion(PFInput *, int64_t delta, bool mouseActive);
void pf_input_fullscreen_done(PFInput *);
uint8_t pf_make_code(uint16_t native);
int64_t pf_clock_ns(uint64_t ticks, uint64_t epoch, uint32_t numer, uint32_t denom);
bool pf_frame_valid(int32_t w, int32_t h, int32_t stride);
void pf_logical_size(int32_t w, int32_t h, double *lw, double *lh);
/* SPSC: main-thread producer; AudioUnit callback consumer. Reset only stopped. */
#define PF_RING_FRAMES 48000u
typedef struct {
    _Atomic uint64_t read, write, underruns, dropped;
    int16_t samples[PF_RING_FRAMES * 2];
} PFRing;
void pf_ring_init(PFRing *);
void pf_ring_reset(PFRing *);
size_t pf_ring_write(PFRing *, const void *, size_t frames);
size_t pf_ring_read(PFRing *, void *, size_t frames);
size_t pf_ring_available(PFRing *);
#endif
