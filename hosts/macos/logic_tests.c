#include "host_logic.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
static struct { PFHostEvent type; int32_t a,b; } events[256];
static size_t count;
static void emit(void *ctx, PFHostEvent t,int32_t a,int32_t b) {
    (void)ctx; assert(count<256); events[count].type=t; events[count].a=a;
    events[count++].b=b;
}
static void input_tests(void) {
    PFInput p; pf_input_init(&p,emit,NULL); pf_input_focus(&p,true);
    /* SNAIL physical makes retain order; repeats and duplicate downs disappear. */
    unsigned keys[]={1,45,0,34,37}; unsigned dos[]={31,49,30,23,38};
    for (unsigned i=0;i<5;i++) {
        pf_input_key(&p,keys[i],true,false,false,false);
        pf_input_key(&p,keys[i],true,true,false,false);
        pf_input_key(&p,keys[i],true,false,false,false);
        pf_input_key(&p,keys[i],false,false,false,false);
    }
    assert(count==5);
    for (unsigned i=0;i<5;i++) assert(events[i].type==PF_EVENT_KEY && events[i].a==(int)dos[i]);
    pf_input_key(&p,49,true,false,true,false);
    bool noModifiers[6]={0}; pf_input_modifiers(&p,noModifiers);
    assert(!p.held[3] && !p.down[49]);
    pf_input_key(&p,49,false,false,false,false); assert(count==5);
    count=0; bool sides[6]={true,false,false,false,false,false};
    pf_input_modifiers(&p,sides); assert(p.held[0] && !p.held[1]);
    sides[1]=true; pf_input_modifiers(&p,sides); assert(p.held[0] && p.held[1]);
    sides[0]=false; pf_input_modifiers(&p,sides); assert(!p.held[0] && p.held[1]);
    sides[5]=true; pf_input_modifiers(&p,sides); assert(p.held[1]);
    sides[1]=false; pf_input_modifiers(&p,sides); assert(p.held[1]);
    pf_input_focus(&p,false); assert(!p.held[1] && !p.fire);
    count=0; pf_input_key(&p,125,true,false,false,false); assert(count==0);
    pf_input_focus(&p,true); assert(!p.down[125]);
    pf_input_key(&p,125,true,false,false,false); assert(p.held[2]);
    pf_input_key(&p,125,false,false,false,false); assert(!p.held[2]);
    count=0;
    pf_input_button(&p,true,false); pf_input_button(&p,true,true); assert(count==0);
    pf_input_button(&p,false,true); pf_input_button(&p,true,true);
    pf_input_button(&p,true,true); assert(count==1 && events[0].type==PF_EVENT_FIRE);
    pf_input_motion(&p,8,true); pf_input_motion(&p,999,false);
    assert(count==2 && events[1].type==PF_EVENT_DELTA && events[1].a==8);
    count=0; pf_input_key(&p,36,true,false,false,true);
    assert(count==1 && events[0].type==PF_EVENT_FULLSCREEN);
    pf_input_key(&p,36,false,false,false,false); pf_input_key(&p,36,true,false,false,true);
    assert(count==1); pf_input_fullscreen_done(&p);
    pf_input_key(&p,36,false,false,false,false); pf_input_key(&p,36,true,false,false,true);
    assert(count==2);
    puts("PASS key mapping, ordered cheat makes, repeat, sided modifiers, focus, fullscreen, plunger/fire");
}
static void *produce(void *ctx) {
    PFRing *r=ctx;
    for (int i=0;i<100000;i++) {
        int16_t pcm[2]={(int16_t)i,(int16_t)~i};
        while (pf_ring_available(r)==PF_RING_FRAMES) { }
        assert(pf_ring_write(r,pcm,1)==1);
    }
    return NULL;
}
static void ring_tests(void) {
    PFRing *r=calloc(1,sizeof(*r)); pf_ring_init(r);
    int16_t pcm[6]={123,-234,456,-567,0,0},out[8];
    assert(pf_ring_write(r,pcm,2)==2); assert(pf_ring_read(r,out,4)==2);
    assert(out[0]==123 && out[1]==-234 && out[2]==456 && out[3]==-567);
    assert(out[4]==0 && out[7]==0 && atomic_load(&r->underruns)==1);
    int16_t *large=calloc(PF_RING_FRAMES*2+2,sizeof(int16_t));
    assert(pf_ring_write(r,large,PF_RING_FRAMES+1)==PF_RING_FRAMES);
    assert(atomic_load(&r->dropped)==1); pf_ring_reset(r); assert(pf_ring_available(r)==0);
    pthread_t thread; assert(!pthread_create(&thread,NULL,produce,r));
    for (int i=0;i<100000;i++) {
        while (!pf_ring_available(r)) { }
        assert(pf_ring_read(r,out,1)==1);
        assert(out[0]==(int16_t)i && out[1]==(int16_t)~i);
    }
    assert(!pthread_join(thread,NULL)); free(large); free(r);
    puts("PASS PCM channel order, silence, capacity, wrap, stopped reset, 100000 concurrent stereo frames");
}
int main(void) {
    assert(pf_clock_ns(100,100,125,3)==0);
    assert(pf_clock_ns(103,100,125,3)==125);
    assert(pf_clock_ns(100,101,1,1)==-1);
    assert(pf_clock_ns(UINT64_MAX,0,UINT32_MAX,1)==INT64_MAX);
    assert(pf_clock_ns(100+3600000000000ULL,100,1,1)==3600000000000LL);
    assert(pf_frame_valid(320,576,1280)); assert(!pf_frame_valid(320,576,1));
    double w,h; pf_logical_size(640,240,&w,&h); assert(w==640 && h==480);
    pf_logical_size(320,576,&w,&h); assert(w==320 && h==576);
    input_tests(); ring_tests(); puts("PASS monotonic conversion and framebuffer geometry");
}
