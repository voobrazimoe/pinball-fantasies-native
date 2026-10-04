package io.github.voobrazimoe.pinballfantasies;

public final class AudioPolicyTest {
    static final class Host implements AudioPolicy.Host {
        int requests, abandons, changes; boolean grant=true, active;
        public boolean request() { requests++; return grant; }
        public void abandon() { abandons++; }
        public void output(boolean value) { if (value != active) changes++; active=value; }
    }
    static void check(boolean value) { if (!value) throw new AssertionError(); }
    public static void main(String[] args) {
        Host h=new Host(); AudioPolicy p=new AudioPolicy(h);
        p.focus(1); check(!h.active && h.requests==0);
        p.eligible(true); check(h.active && h.requests==1);
        p.eligible(true); p.focus(1); check(h.requests==1 && h.changes==1); // rotation / duplicate gain
        p.focus(-2); p.focus(-2); check(!h.active && h.changes==2);
        p.focus(1); check(h.active && h.changes==3);
        p.focus(-3); check(!h.active); p.focus(1); check(h.active);
        p.focus(-1); p.focus(1); p.eligible(true); check(!h.active && h.requests==1);
        p.eligible(false); p.focus(1); p.eligible(true); check(h.active && h.requests==2);
        p.eligible(false); p.focus(-2); p.focus(1); check(!h.active);
        p.eligible(true); p.focus(-2); p.eligible(false); p.eligible(true); check(h.active);
        p.close(); p.focus(1); p.eligible(true); check(!h.active);
        Host fresh=new Host(); AudioPolicy next=new AudioPolicy(fresh);
        next.eligible(true); p.focus(-1); check(fresh.active); // old Activity listener
        fresh.grant=false; next.eligible(false); next.eligible(true);
        int attempts=fresh.requests; next.eligible(true); next.focus(1);
        check(!fresh.active && fresh.requests==attempts); // failure stays silent
        fresh.grant=true; next.eligible(false); next.eligible(true); check(fresh.active);
        next.close(); System.out.println("PASS: A5 focus, duck, permanent loss, failure, lifecycle, rotation and stale listener reducer");
    }
}
