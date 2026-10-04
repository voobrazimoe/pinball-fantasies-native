package io.github.voobrazimoe.pinballfantasies;
import java.util.ArrayList;
import java.util.List;
public final class ControlsTest {
    static void check(boolean b) { if (!b) throw new AssertionError(); }
    public static void main(String[] args) throws Exception {
        if (args.length>0) try (java.util.Scanner mappings=new java.util.Scanner(new java.io.File(args[0]))) {
            while(mappings.hasNextInt()) {
                int key=mappings.nextInt(), expected=mappings.nextInt();
                if (Controls.make(key)!=expected) throw new AssertionError("Desktop make parity for Android key "+key);
            }
        }
        List<String> events=new ArrayList<>();
        List<Runnable> releases=new ArrayList<>();
        Controls c=new Controls((kind,a,b)->events.add(kind+":"+a+":"+b),
                (ms,release)->{ check(ms==50); releases.add(release); }); c.enabled=true;
        c.geometry(400,800,0,0,0,0,44,8,500);
        check(c.hit(100,700)==Controls.LEFT && c.hit(300,700)==Controls.RIGHT);
        check(c.hit(399,700)==Controls.RIGHT); // No bottom plunger region.
        check(c.hit(100,20)==Controls.NONE && c.hit(-1,700)==Controls.NONE);
        c.down(17,100,700,0); c.down(91,300,700,0);
        check(c.held[0] && c.held[1]);
        c.up(91,100); check(c.held[0] && !c.held[1]);
        c.down(37,100,700,0); c.up(17,100); check(c.held[0]);
        c.key(59,true,0); c.up(37,100); check(c.held[0]);
        c.key(113,true,0); c.key(59,false,0); check(c.held[0]);
        c.key(113,false,0); check(!c.held[0]);
        c.down(23,100,700,0); c.down(8,300,700,0); c.cancel();
        check(!c.held[0] && !c.held[1] && c.pointers.isEmpty());
        for (float x:new float[]{100,300}) {
            events.clear(); c.down(7,x,250,0); check(events.isEmpty());
            c.move(7,x+2,254); check(events.isEmpty()); c.up(7,100); c.up(7,100);
            check(events.equals(List.of("0:3:1","1:57:0")));
            releases.remove(0).run(); check(!c.held[3]); // Exactly one release nudge.
        }
        events.clear(); c.down(7,100,250,0); c.move(7,100,270); c.up(7,100);
        check(events.isEmpty()); // Central drag is not a tap.
        c.down(7,100,250,0); c.up(7,501); check(events.isEmpty());
        c.down(7,300,250,0); c.cancel(); check(events.isEmpty());
        c.down(7,300,250,0); c.move(7,300,258); check(!c.pulling()); c.up(7,100);
        check(events.equals(List.of("0:3:1","1:57:0")));
            releases.remove(0).run(); check(!c.held[3]); // Threshold is strictly exceeded.
        events.clear(); c.down(7,300,250,0); c.down(8,310,260,0);
        c.move(7,300,260); check(c.pulling());
        c.move(8,310,280); check(c.pointers.get(8).region==Controls.NONE);
        c.up(8,100); // Other spring candidate cannot fire or nudge.
        c.move(7,310,250+c.pullTravel/2); c.move(7,310,250+c.pullTravel/2);
        check(c.pointers.get(7).position==64);
        c.move(7,100,250+c.pullTravel); // Horizontal jitter/crossing keeps ownership.
        check(c.pointers.get(7).position==128 && c.pulling());
        check(c.pullTravel<200); // Full spring within practical portrait travel.
        check(events.stream().filter(e->e.startsWith("3:")).mapToInt(e->Integer.parseInt(e.split(":")[1])).sum()==128);
        c.up(7,400); c.up(7,400); check(!c.pulling());
        check(events.stream().filter(e->e.equals("4:0:0")).count()==1);
        check(!events.contains("1:57:0"));
        c.down(9,300,250,0); c.move(9,300,280); events.clear(); c.cancel();
        check(events.equals(List.of("4:0:0")) && !c.pulling());
        c.down(1,100,700,0); events.clear(); c.down(2,100,250,0); c.up(2,100);
        check(c.held[0] && events.equals(List.of("0:3:1","1:57:0")));
        releases.remove(0).run(); check(c.held[0] && !c.held[3]); c.up(1,100);
        for (float x:new float[]{100,300}) {
            c.down(1,x,700,0); c.down(2,300,250,0); c.move(2,300,300);
            check(c.pulling() && c.held[x<200?0:1]);
            c.up(2,100); check(c.held[x<200?0:1]); c.up(1,100);
            c.down(1,x,700,0); c.down(2,300,250,0); c.move(2,300,300);
            c.up(1,100); check(c.pulling()); c.cancel(); check(!c.pulling());
        }
        // Tall, short portrait and landscape; synthetic asymmetric cutout/bar edges.
        for (float[] shape:new float[][]{{400,900,0,80,0,24},{360,640,12,48,24,16},{900,400,80,0,40,24}}) {
            float w=shape[0],h=shape[1],l=shape[2],t=shape[3],r=shape[4],b=shape[5];
            c.geometry(w,h,l,t,r,b,t+44,8,500);
            check(c.hit(l-1,h/2)==Controls.NONE && c.hit(w-r,h/2)==Controls.NONE);
            check(c.hit(w/2,t+43)==Controls.NONE && c.hit(w/2,h-b)==Controls.NONE);
            check(c.hit(l+1,h-b-1)==Controls.LEFT && c.hit(w-r-1,h-b-1)==Controls.RIGHT);
            c.down(3,w-r-1,c.top+10,0); c.move(3,w-r-1,c.top+10+c.pullTravel);
            check(c.pulling() && c.pointers.get(3).position==128); c.cancel();
        }
        events.clear(); c.down(9,850,100,0); c.move(9,850,100+c.pullTravel);
        check(c.pulling()); c.move(9,10,100); check(c.pointers.get(9).position==0);
        check(events.equals(List.of("3:128:0","3:-128:0")));
        c.down(10,850,100,0); check(!c.pointers.containsKey(10));
        c.up(10,100); check(c.pulling()); events.clear(); c.clear();
        check(!c.pulling() && events.isEmpty()); // Focus suspend discards spring, no fire.
        c.down(9,200,100,0); c.up(9,100); check(c.held[3]);
        c.key(62,true,0); releases.remove(0).run(); check(c.held[3]); // Keyboard owns Tilt too.
        c.key(62,false,0); check(!c.held[3]);
        c.down(9,200,100,0); c.up(9,100); c.clear(); events.clear();
        releases.remove(0).run(); check(events.isEmpty()); // Stale pulse after suspend.
        for(int k=131;k<=138;k++) check(Controls.make(k)==59+k-131);
        check(Controls.make(4)==1 && Controls.make(111)==1);
        check(Controls.make(66)==28 && Controls.make(20)==80);
        check(Controls.make(44)==25 && Controls.make(41)==50 && Controls.make(62)==57);
        events.clear(); c.key(66,true,0); c.key(66,true,1); c.key(66,true,0); c.key(66,false,0);
        check(events.equals(List.of("2:0:0","1:28:0")));
        c.key(20,true,0); check(c.held[2]); c.key(20,false,0); check(!c.held[2]);
        c.down(1,100,300,0); c.key(60,true,0); c.clear(); c.enabled=false;
        c.up(1,100); c.key(60,false,0); c.down(1,100,300,0);
        check(!c.held[0] && !c.held[1] && c.pointers.isEmpty());
        c.enabled=true; c.key(4,true,0); c.key(4,false,0);
        check(events.get(events.size()-1).equals("1:1:0"));
        System.out.println("PASS: A3 pointer ownership, flippers, cancellation, plunger, keyboard and Back");
    }
}
