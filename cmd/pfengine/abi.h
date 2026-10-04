#ifndef PF_ENGINE_ABI_H
#define PF_ENGINE_ABI_H
#include <stdint.h>
#include <stddef.h>
#ifdef __cplusplus
extern "C" {
#endif
#define PF_ABI_VERSION 1
#define PF_OK 0
#define PF_INVALID -1
#define PF_BUSY -2
#define PF_ERROR -3
#define PF_LEFT 0
#define PF_RIGHT 1
#define PF_SPRING 2
#define PF_TILT 3
#define PF_MODE_STARTUP 0
#define PF_MODE_SELECTOR 1
#define PF_MODE_SELECTOR_TEXT 2
#define PF_MODE_OPTIONS 3
#define PF_MODE_ATTRACT 4
#define PF_MODE_PLAYING 5
#define PF_MODE_PAUSED 6
#define PF_MODE_QUIT_QUESTION 7
#define PF_MODE_GAME_END 8
#define PF_MODE_INITIALS 9
#define PF_MODE_ENTRY_WAIT 10
#define PF_MODE_QUIT 11
/* Sink runs synchronously on the engine call path. PCM is interleaved signed
 * little-endian int16, 48000 Hz stereo; bytes is divisible by four. Borrowed
 * samples are valid ONLY during the callback: copy into a host-owned ring.
 * No engine reentry; native device callbacks consume the ring only. */
typedef void (*pf_pcm_sink)(void *context, const uint8_t *samples, uint32_t bytes);
int32_t pf_engine_abi_version(void);
/* data/state are UTF-8 paths, copied during creation. Error receives a bounded
 * NUL-terminated message when capacity > 0. Zero handle means failure. */
uint64_t pf_engine_create(char *data, char *state, int64_t ns, char *error, uint32_t capacity);
int32_t pf_engine_destroy(uint64_t handle);
int32_t pf_engine_suspend(uint64_t handle);
int32_t pf_engine_resume(uint64_t handle, int64_t ns);
int32_t pf_engine_set_action(uint64_t handle, uint32_t action, int32_t down);
int32_t pf_engine_key(uint64_t handle, uint8_t logical_make);
int32_t pf_engine_release(uint64_t handle);
int32_t pf_engine_plunger_delta(uint64_t handle, int32_t delta);
/* Optional additive ABI 1 touch extension: clamped absolute 0..32 charge,
 * ignored outside active valid spring input. Consumed on the source task.
 * Existing relative DOS mouse semantics and release timing are unchanged. */
int32_t pf_engine_plunger_target(uint64_t handle, int32_t target);
int32_t pf_engine_plunger_fire(uint64_t handle);
int32_t pf_engine_advance(uint64_t handle, int64_t ns, pf_pcm_sink sink, void *context);
/* RGBA8, top-down, engine-owned C storage. Valid until next frame retrieval or
 * destroy. Do not free/write. Retrieval copies the authoritative Go raster into
 * reusable C storage: never exports a retained Go pointer; no per-frame malloc.
 * Every out pointer must be non-null. */
/* Optional additive ABI 1 extension. full_table must be 0 (saved presentation)
 * or 1 (full 320x609 table presentation). Changes only frame composition;
 * never settings, simulation, cadence, audio, table or player state. Serialized
 * like all other calls. Older hosts need not call it; default is 0. */
int32_t pf_engine_set_presentation(uint64_t handle, int32_t full_table);
int32_t pf_engine_frame(uint64_t handle, uint8_t **pixels, int32_t *width, int32_t *height, int32_t *stride);
/* flags: bit 0 suspended, bit 1 quit, bit 2 mouse plunger active. */
int32_t pf_engine_state(uint64_t handle, uint64_t *tick, uint32_t *mode, uint32_t *table, uint32_t *flags);
#ifdef __cplusplus
}
#endif
#endif
