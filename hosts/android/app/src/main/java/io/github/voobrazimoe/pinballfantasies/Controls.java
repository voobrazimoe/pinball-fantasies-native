package io.github.voobrazimoe.pinballfantasies;

import java.util.HashMap;
import java.util.Map;

// Android-independent translation policy, used by the overlay and hardware keys.
final class Controls {
    interface Sink { void send(int kind, int a, int b); }
    final Sink sink;
    final Map<Integer, Pointer> pointers = new HashMap<>();
    final boolean[] keys = new boolean[256], held = new boolean[4];
    boolean enabled;
    static final int LEFT=0, RIGHT=1, PLUNGER=4, NUDGE=3, NONE=-1;
    static final class Pointer {
        final int region;
        final float origin;
        int position;
        Pointer(int region, float y) { this.region=region; origin=y; }
    }
    Controls(Sink sink) { this.sink=sink; }
    // Bottom 35%: left 40%, right 40%, plunger rightmost 20%.
    // Middle right edge: nudge. Top 20% is reserved for the menu overlay.
    static int hit(float x, float y) {
        if (x<0 || x>1 || y<0 || y>1) return NONE;
        if (y>=.65f) return x<.4f ? LEFT : x<.8f ? RIGHT : PLUNGER;
        return x>=.8f && y>=.2f ? NUDGE : NONE;
    }
    void down(int id, float x, float y) {
        if (!enabled || pointers.containsKey(id)) return;
        int region=hit(x,y);
        if (region==PLUNGER) {
            for (Pointer p:pointers.values()) if (p.region==PLUNGER) return;
        }
        pointers.put(id,new Pointer(region,y));
        actions();
        // Like physical makes, touch presses can dismiss startup/pause prompts.
        if (region==LEFT || region==RIGHT) sink.send(1,127,0);
        if (region==NUDGE) sink.send(1,57,0);
    }
    void move(int id, float y) {
        Pointer p=pointers.get(id);
        if (!enabled || p==null || p.region!=PLUNGER) return;
        // Pull downward up to 25% of screen height = 128 relative counts.
        // Absolute quantization makes total delta independent of callback count.
        int position=Math.round(Math.max(0,Math.min(1,(y-p.origin)/.25f))*128);
        int delta=position-p.position;
        p.position=position;
        if (delta!=0) sink.send(3,delta,0);
    }
    void up(int id) {
        Pointer p=pointers.remove(id);
        if (p==null) return;
        if (p.region==PLUNGER) sink.send(4,0,0);
        actions();
    }
    void cancel() {
        boolean plunger=false;
        for (Pointer p:pointers.values()) plunger |= p.region==PLUNGER;
        pointers.clear();
        // Existing mouse fire releases a charged spring; no host physics.
        if (plunger) sink.send(4,0,0);
        actions();
    }
    void clear() {
        pointers.clear();
        java.util.Arrays.fill(keys,false);
        java.util.Arrays.fill(held,false);
        // Native suspend clears both pending mouse deltas and all held actions.
    }
    void actions() {
        boolean[] next=new boolean[4];
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
        if (!enabled) return;
        if (code==28) sink.send(2,0,0);
        sink.send(1,code,0);
    }
}
