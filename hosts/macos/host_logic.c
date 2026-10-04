#include "host_logic.h"
#include <limits.h>
#include <string.h>

uint8_t pf_make_code(uint16_t k) {
    switch (k) {
    case 53: return 1; case 36: case 76: return 28; case 49: return 57;
    case 122: return 59; case 120: return 60; case 99: return 61; case 118: return 62;
    case 96: return 63; case 97: return 64; case 98: return 65; case 100: return 66;
    case 126: return 72; case 125: return 80; case 67: return 55;
    case 12: return 16; case 13: return 17; case 14: return 18; case 15: return 19;
    case 17: return 20; case 16: return 21; case 32: return 22; case 34: return 23;
    case 31: return 24; case 35: return 25; case 0: return 30; case 1: return 31;
    case 2: return 32; case 3: return 33; case 5: return 34; case 4: return 35;
    case 38: return 36; case 40: return 37; case 37: return 38; case 6: return 44;
    case 7: return 45; case 8: return 46; case 9: return 47; case 11: return 48;
    case 45: return 49; case 46: return 50;
    default: return 127;
    }
}
static void event(PFInput *p, PFHostEvent e, int32_t a, int32_t b) {
    if (p->emit) p->emit(p->context, e, a, b);
}
static void holds(PFInput *p) {
    bool h[4] = {p->down[56] || p->down[59] || p->down[58],
                 p->down[60] || p->down[62] || p->down[61],
                 p->down[125], p->down[49]};
    for (int i=0;i<4;i++) if (h[i]!=p->held[i]) {
        p->held[i]=h[i]; event(p,PF_EVENT_ACTION,i,h[i]);
    }
}
void pf_input_init(PFInput *p, PFEmit emit, void *context) {
    memset(p,0,sizeof(*p)); p->emit=emit; p->context=context;
}
void pf_input_key(PFInput *p, uint16_t k, bool down, bool repeat, bool command, bool option) {
    if (!p->focused || k>=128 || repeat) return;
    if (p->down[k]==down) return;
    bool fullscreen=(k==36 && option) || (k==3 && command);
    if (command && down && !fullscreen) return;
    p->down[k]=down;
    if (down && fullscreen) {
        if (!p->fullscreenPending) {
            p->fullscreenPending=true; event(p,PF_EVENT_FULLSCREEN,0,0);
        }
        return;
    }
    /* Command menu shortcuts never become gameplay contributors. */
    holds(p);
    if (down) {
        if (k==36 || k==76) event(p,PF_EVENT_RELEASE,0,0);
        event(p,PF_EVENT_KEY,pf_make_code(k),0);
    }
}
void pf_input_modifiers(PFInput *p, const bool sides[6]) {
    if (!p->focused) return;
    const unsigned keys[6]={56,60,59,62,58,61};
    unsigned makes=0;
    for (unsigned i=0;i<6;i++) {
        if (sides[i] && !p->down[keys[i]]) makes++;
        p->down[keys[i]]=sides[i];
    }
    holds(p);
    /* Modifier-only native events must also reach shared pause/quit semantics,
       as ordinary desktop modifier makes do. They are never auto-repeat. */
    while (makes--) event(p,PF_EVENT_KEY,127,0);
}
void pf_input_focus(PFInput *p, bool focused) {
    if (!focused) {
        memset(p->down,0,sizeof(p->down));
        /* Suspend is sent first by the host: these zeros cannot create a spring edge. */
        for (int i=0;i<4;i++) if (p->held[i]) {
            p->held[i]=false; event(p,PF_EVENT_ACTION,i,0);
        }
        p->fire=false;
    }
    p->focused=focused;
}
void pf_input_button(PFInput *p, bool down, bool active) {
    if (p->focused && active && down && !p->fire) event(p,PF_EVENT_FIRE,0,0);
    p->fire=down;
}
void pf_input_motion(PFInput *p, int64_t delta, bool active) {
    if (!p->focused || !active) return;
    if (delta>INT32_MAX) delta=INT32_MAX;
    if (delta<INT32_MIN) delta=INT32_MIN;
    if (delta) event(p,PF_EVENT_DELTA,(int32_t)delta,0);
}
void pf_input_fullscreen_done(PFInput *p) { p->fullscreenPending=false; }
int64_t pf_clock_ns(uint64_t ticks, uint64_t epoch, uint32_t numer, uint32_t denom) {
    if (!denom || ticks<epoch) return -1;
    __uint128_t ns=(__uint128_t)(ticks-epoch)*numer/denom;
    return ns>INT64_MAX ? INT64_MAX : (int64_t)ns;
}
bool pf_frame_valid(int32_t w, int32_t h, int32_t stride) {
    return w>0 && h>0 && w<=4096 && h<=4096 && stride>=w*4 && stride<=4096*4;
}
void pf_logical_size(int32_t w,int32_t h,double *lw,double *lh) {
    *lw=w; *lh=(w==640 && h==240)?480:h;
}
void pf_ring_init(PFRing *r) {
    atomic_init(&r->read,0); atomic_init(&r->write,0);
    atomic_init(&r->underruns,0); atomic_init(&r->dropped,0);
}
void pf_ring_reset(PFRing *r) {
    atomic_store(&r->read,0); atomic_store(&r->write,0);
}
size_t pf_ring_available(PFRing *r) {
    return (size_t)(atomic_load_explicit(&r->write,memory_order_acquire)-
                    atomic_load_explicit(&r->read,memory_order_acquire));
}
size_t pf_ring_write(PFRing *r,const void *pcm,size_t frames) {
    uint64_t w=atomic_load_explicit(&r->write,memory_order_relaxed);
    uint64_t rd=atomic_load_explicit(&r->read,memory_order_acquire);
    size_t free=PF_RING_FRAMES-(size_t)(w-rd), n=frames<free?frames:free;
    const unsigned char *src=pcm;
    for (size_t i=0;i<n;i++) memcpy(&r->samples[((w+i)%PF_RING_FRAMES)*2],src+i*4,4);
    atomic_store_explicit(&r->write,w+n,memory_order_release);
    atomic_fetch_add(&r->dropped,frames-n);
    return n;
}
size_t pf_ring_read(PFRing *r,void *pcm,size_t frames) {
    uint64_t rd=atomic_load_explicit(&r->read,memory_order_relaxed);
    uint64_t w=atomic_load_explicit(&r->write,memory_order_acquire);
    size_t available=(size_t)(w-rd), n=frames<available?frames:available;
    unsigned char *dst=pcm;
    for (size_t i=0;i<n;i++) memcpy(dst+i*4,&r->samples[((rd+i)%PF_RING_FRAMES)*2],4);
    memset(dst+n*4,0,(frames-n)*4);
    atomic_store_explicit(&r->read,rd+n,memory_order_release);
    if (n<frames) atomic_fetch_add(&r->underruns,1);
    return n;
}
