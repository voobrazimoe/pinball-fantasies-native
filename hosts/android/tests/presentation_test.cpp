#include "presentation_policy.h"
#include "presentation_diagnostics.h"
#include <cassert>
int main() {
    presentation::Policy p;
    assert(p.pollTimeoutMillis()==-1 && !p.arm() && !p.consume());
    p.setActive(true); // resume + window + focus
    assert(p.pollTimeoutMillis()==-1 && p.arm() && !p.arm());
    assert(p.consume() && !p.consume()); // at most one presentation per callback
    assert(p.arm());
    p.setActive(false); // pause / focus loss / TERM_WINDOW
    assert(!p.arm() && !p.consume() && !p.arm());
    p.setActive(true);
    assert(p.arm());
    p.setActive(false);
    p.setActive(true); // replacement window before old callback arrives
    assert(!p.arm() && !p.consume()); // discard old epoch
    assert(p.arm() && p.consume());
    assert(p.arm());
    p.setActive(false); // destruction drains without presentation or re-arm
    assert(p.pending() && !p.consume() && !p.pending() && !p.arm());
    // No callback timestamp or counter enters the policy's eligibility rules.
    // Interval aggregation follows actual times, including 120 Hz and missed vsync.
    presentation::Diagnostics d;
    d.begin(1000000000,1000000000);
    d.finish(1001000000,true,0);
    d.begin(1008333333,1008333333);
    d.finish(1009333333,true,1);
    d.begin(1024999999,1024999999);
    d.finish(1025999999,true,2);
    assert(d.attempts==3 && d.swaps==3 && d.ticks==3 && d.knownTicks==3);
    assert(d.zero==1 && d.one==1 && d.many==1);
    assert(d.vsyncIntervals.count==2 && d.vsyncIntervals.min==8333333);
    d.begin(1041666666,1041666666);
    d.finish(1042666666,false,4);
    assert(d.attempts==4 && d.swaps==3 && d.ticks==3);
}
