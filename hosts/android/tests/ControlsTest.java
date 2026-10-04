package io.github.voobrazimoe.pinballfantasies;
import java.util.ArrayList;
import java.util.List;
public final class ControlsTest {
    static void check(boolean b) { if (!b) throw new AssertionError(); }
    public static void main(String[] args) {
        List<String> events=new ArrayList<>();
        Controls c=new Controls((kind,a,b)->events.add(kind+":"+a+":"+b)); c.enabled=true;
        check(Controls.hit(.1f,.8f)==0 && Controls.hit(.6f,.8f)==1);
        check(Controls.hit(.9f,.8f)==4 && Controls.hit(.9f,.4f)==3);
        check(Controls.hit(.1f,.1f)==-1 && Controls.hit(-1,.8f)==-1);
        c.down(17,.1f,.8f); c.down(91,.6f,.8f);
        check(c.held[0] && c.held[1]);
        c.up(91); check(c.held[0] && !c.held[1]);
        c.down(37,.2f,.8f); c.up(17); check(c.held[0]);
        c.key(59,true,0); c.up(37); check(c.held[0]);
        c.key(113,true,0); c.key(59,false,0); check(c.held[0]);
        c.key(113,false,0); check(!c.held[0]);
        c.down(23,.1f,.8f); c.down(8,.6f,.8f); c.cancel();
        check(!c.held[0] && !c.held[1] && c.pointers.isEmpty());
        c.down(7,.9f,.7f); c.down(8,.9f,.72f);
        check(c.pointers.size()==1); // One plunger owner; unrelated pointer up is harmless.
        events.clear(); c.move(7,.825f); c.move(7,.825f); c.up(8);
        check(events.equals(List.of("3:64:0")));
        c.move(7,2); c.move(7,.7f); c.up(7);
        check(events.equals(List.of("3:64:0","3:64:0","3:-128:0","4:0:0")));
        c.down(9,.9f,.7f); c.move(9,.8f); events.clear(); c.cancel();
        check(events.equals(List.of("4:0:0")));
        for(int k=131;k<=138;k++) check(Controls.make(k)==59+k-131);
        check(Controls.make(4)==1 && Controls.make(111)==1);
        check(Controls.make(66)==28 && Controls.make(20)==80);
        check(Controls.make(44)==25 && Controls.make(41)==50 && Controls.make(62)==57);
        events.clear(); c.key(66,true,0); c.key(66,true,1); c.key(66,true,0); c.key(66,false,0);
        check(events.equals(List.of("2:0:0","1:28:0")));
        c.key(20,true,0); check(c.held[2]); c.key(20,false,0); check(!c.held[2]);
        c.down(1,.1f,.8f); c.key(60,true,0); c.clear(); c.enabled=false;
        c.up(1); c.key(60,false,0); c.down(1,.1f,.8f);
        check(!c.held[0] && !c.held[1] && c.pointers.isEmpty());
        c.enabled=true; c.key(4,true,0); c.key(4,false,0);
        check(events.get(events.size()-1).equals("1:1:0"));
        System.out.println("PASS: A3 pointer ownership, flippers, cancellation, plunger, keyboard and Back");
    }
}
