package io.github.voobrazimoe.pinballfantasies;

import java.util.HashMap;
import java.util.Map;

// Android-independent translation policy, used by the overlay and hardware keys.
final class Controls {
    interface Sink { void send(int kind, int a, int b); }
    interface Delay { void after(long milliseconds, Runnable release); }
    static final long NUDGE_PRESS_MS=50;
    private final Delay delay;
    private boolean touchNudge;
    private int nudgeGeneration;
    final Sink sink;
    final Map<Integer, Pointer> pointers = new HashMap<>();
    final boolean[] keys = new boolean[256], held = new boolean[4];
    boolean enabled;
    boolean gameplay=true, plungerAvailable;
    static final int LEFT=0, RIGHT=1, PLUNGER=4, NUDGE=3, NONE=-1, PENDING=5;
    Runnable changed = () -> {};
    // Pixel coordinates supplied by the measured overlay; no device dimensions.
    float left, top, right, bottom, stripTop, slop, pullTravel;
    long tapTimeout;
    static final class Pointer {
        int region;
        final float x, y, travel;
        final long started;
        final boolean eligible;
        boolean tap=true;
        int position;
        Pointer(int region, float x, float y, long time, boolean eligible, float travel) {
            this.region=region; this.x=x; this.y=y; started=time;
            this.eligible=eligible; this.travel=travel;
        }
    }
    Controls(Sink sink, Delay delay) { this.sink=sink; this.delay=delay; }
    void geometry(float width, float height, float safeLeft, float safeTop,
                  float safeRight, float safeBottom, float touchSlop, long timeout) {
        left=safeLeft; right=width-safeRight; top=safeTop;
        bottom=height-safeBottom;
        stripTop=top+Math.max(0,bottom-top)*.65f;
        slop=touchSlop; tapTimeout=timeout;
        pullTravel=Math.max(2*slop,(bottom-top)*.25f);
    }
    float labelX(int region) { return left+(right-left)*(region==LEFT ? .25f : .75f); }
    float labelY() { return (stripTop+bottom)/2; }
    int hit(float x, float y) {
        if (x<left || x>=right || y<top || y>=bottom || top>=bottom) return NONE;
        if (y>=stripTop) return x<(left+right)/2 ? LEFT : RIGHT;
        return PENDING;
    }
    boolean pulling() {
        for (Pointer p:pointers.values()) if (p.region==PLUNGER) return true;
        return false;
    }
    void down(int id, float x, float y, long time) {
        if (!enabled || !gameplay || pointers.containsKey(id)) return;
        int region=hit(x,y);
        if (region==NONE) return;
        // Starts during another spring owner's gesture cannot become taps/pulls.
        if (region==PENDING && pulling()) return;
        pointers.put(id,new Pointer(region,x,y,time,plungerAvailable && x>=(left+right)/2,pullTravel));
        actions();
        if (region==LEFT || region==RIGHT) sink.send(1,127,0);
        changed.run();
    }
    void move(int id, float x, float y) {
        Pointer p=pointers.get(id);
        if (!enabled || p==null) return;
        float dx=x-p.x, dy=y-p.y;
        if (dx*dx+dy*dy>slop*slop) p.tap=false;
        if (p.region==PENDING && p.eligible && plungerAvailable && dy>slop && dy>Math.abs(dx)) {
            // Resolve at classification, not down: only one spring owner.
            p.region=pulling() ? NONE : PLUNGER;
            if (p.region==PLUNGER) for (Pointer other:pointers.values())
                if (other!=p && other.region==PENDING) other.region=NONE;
        }
        if (p.region==PLUNGER) {
            // Absolute charge follows finger position, independent of callback count.
            int position=Math.round(Math.max(0,Math.min(1,dy/p.travel))*32);
            int delta=position-p.position;
            p.position=position;
            if (delta!=0) sink.send(5,position,0);
        }
        changed.run();
    }
    void up(int id, long time) {
        Pointer p=pointers.remove(id);
        if (p==null) return;
        if (p.region==PLUNGER) sink.send(4,0,0);
        if (p.region==PENDING && p.tap && time-p.started<=tapTimeout && !pulling())
            nudge(); // One momentary Tilt press and Space make, only on release.
        actions(); changed.run();
    }
    private void nudge() {
        touchNudge=true; actions(); sink.send(1,57,0);
        final int generation=++nudgeGeneration;
        // Input press duration only; source cadence/physics remain engine-owned.
        delay.after(NUDGE_PRESS_MS,()->{
            if (generation!=nudgeGeneration) return;
            touchNudge=false; actions();
        });
    }
    void cancel() {
        boolean plunger=pulling();
        pointers.clear();
        touchNudge=false; nudgeGeneration++;
        // Cancel resets the touch target without scheduling a release.
        if (plunger) sink.send(5,0,0);
        actions(); changed.run();
    }
    void clear() {
        pointers.clear();
        touchNudge=false; nudgeGeneration++;
        java.util.Arrays.fill(keys,false);
        java.util.Arrays.fill(held,false);
        changed.run();
        // Native suspend clears both pending mouse deltas and all held actions.
    }
    void actions() {
        boolean[] next=new boolean[4];
        next[NUDGE]=touchNudge;
        for (Pointer p:pointers.values()) if (p.region>=0 && p.region<4) next[p.region]=true;
        for (int k=0;k<keys.length;k++) if (keys[k]) {
            int a=action(k); if (a>=0) next[a]=true;
        }
        for (int a=0;a<4;a++) if (next[a]!=held[a]) {
            held[a]=next[a]; sink.send(0,a,next[a]?1:0);
        }
    }
    static int action(int k) {
        switch(k) {
            case 59: case 113: case 57: case 21: case 54: return LEFT;
            case 60: case 114: case 58: case 22: case 76: return RIGHT;
            case 20: return 2;
            case 62: return NUDGE;
            default: return -1;
        }
    }
    // Android KeyEvent numeric constants, intentionally framework-free for tests.
    // Logical DOS makes match frontend.Key and desktop host translations.
    static int make(int k) {
        if (k>=131 && k<=138) return 59+k-131; // F1-F8
        switch(k) {
            case 4: case 111: return 1; // Back/Escape
            case 66: case 160: return 28;
            case 19: return 72;
            case 20: return 80;
            case 62: return 57;
            case 155: return 55; // numpad multiply
            case 29: return 30; case 30: return 48; case 31: return 46;
            case 32: return 32; case 33: return 18; case 34: return 33;
            case 35: return 34; case 36: return 35; case 37: return 23;
            case 38: return 36; case 39: return 37; case 40: return 38;
            case 41: return 50; case 42: return 49; case 43: return 24;
            case 44: return 25; case 45: return 16; case 46: return 19;
            case 47: return 31; case 48: return 20; case 49: return 22;
            case 50: return 47; case 51: return 17; case 52: return 45;
            case 53: return 21; case 54: return 44;
            default: return action(k)>=0 ? 127 : -1;
        }
    }
    boolean key(int k, boolean down, int repeats) {
        int code=make(k);
        if (code<0) return false;
        if (!enabled || k<0 || k>=keys.length) return true;
        if ((down && repeats!=0) || keys[k]==down) return true;
        keys[k]=down;
        actions();
        if (down) {
            if (code==28) sink.send(2,0,0);
            sink.send(1,code,0);
        }
        return true;
    }
    void tap(int code) {
        if (!enabled || code<0) return;
        if (code==28) sink.send(2,0,0);
        sink.send(1,code,0);
    }
}
