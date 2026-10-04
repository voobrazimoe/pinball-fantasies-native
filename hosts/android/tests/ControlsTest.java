package io.github.voobrazimoe.pinballfantasies;
import java.util.*;
public final class ControlsTest {
    static void check(boolean b) { if(!b) throw new AssertionError(); }
    public static void main(String[] args) throws Exception {
        if(args.length>0) try(Scanner s=new Scanner(new java.io.File(args[0]))) {
            while(s.hasNextInt()) check(Controls.make(s.nextInt())==s.nextInt());
        }
        List<String> events=new ArrayList<>(); List<Runnable> releases=new ArrayList<>();
        Controls c=new Controls((k,a,b)->events.add(k+":"+a+":"+b),(ms,r)->{check(ms==50);releases.add(r);});
        c.enabled=true; c.plungerAvailable=true;
        for(int[] shape:new int[][]{{1080,2400,0,172,1080,2055},{2400,1080,480,0,1440,1080},{1920,1080,240,0,1440,1080}}) {
            c.geometry(shape[0],shape[1],24,48,36,48,24,500,3,
                    new InteractionGeometry.Rect(shape[2],shape[3],shape[2]+shape[4],shape[3]+shape[5]));
            InteractionGeometry g=c.layout;
            float lx=g.left.cx(),ly=g.left.cy(),rx=g.right.cx(),ry=g.right.cy();
            check(c.labelX(Controls.LEFT)==lx && c.labelY(Controls.LEFT)==ly);
            check(c.labelX(Controls.RIGHT)==rx && c.labelY(Controls.RIGHT)==ry);
            check(c.region(Controls.LEFT)==g.left && c.region(Controls.RIGHT)==g.right); // Drawing uses hit objects.
            c.down(17,lx,ly,0); c.down(91,rx,ry,0); check(c.held[0] && c.held[1]);
            c.up(17,100); check(!c.held[0] && c.held[1]); c.cancel(); check(!c.held[1]);
            c.down(17,lx,ly,0); c.down(18,lx,ly,0); c.up(17,100); check(c.held[0]);
            c.key(59,true,0); c.up(18,100); check(c.held[0]); c.key(59,false,0); check(!c.held[0]);
            events.clear();
            for(InteractionGeometry.Rect zone:new InteractionGeometry.Rect[]{g.left,g.right}) {
                c.down(4,zone.cx(),zone.top-20,0); c.move(4,zone.cx(),zone.cy()); c.up(4,100);
                check(events.isEmpty()); // Guard starts stay dead; no surprise nudge or delayed flipper.
            }
            c.down(1,lx,ly,0); c.down(2,rx,ry,0);
            float px=g.plunger.cx(),py=g.plunger.top+10,nx=g.nudge.cx(),ny=g.nudge.cy();
            for(int moves:new int[]{2,20}) {
                events.clear(); c.down(9,px,py,0);
                for(int i=1;i<=moves;i++) c.move(9,px,py+c.pullTravel*i/moves);
                check(c.pulling() && c.pointers.get(9).position==32 && c.held[0] && c.held[1]);
                c.down(8,nx,ny,0); c.up(8,100); c.up(8,100);
                check(events.stream().filter(e->e.equals("1:57:0")).count()==1);
                releases.remove(0).run();
                c.up(9,100); c.up(9,100);
                check(events.stream().filter(e->e.equals("4:0:0")).count()==1);
            }
            c.up(1,100); c.up(2,100);
            events.clear(); c.down(9,px,py,0); c.down(10,px+2,py,0);
            c.move(9,px,py+25); c.move(10,px,py+100);
            check(c.pulling() && c.pointers.get(10).region==Controls.NONE);
            c.up(10,100); check(!events.contains("4:0:0"));
            c.move(9,nx,py+c.pullTravel); check(c.pointers.get(9).position==32);
            c.cancel();

            c.down(9,px,py,0); c.move(9,px,py+c.pullTravel/2); check(c.pointers.get(9).position==16);
            c.move(9,px,py); check(c.pointers.get(9).position==0);
            c.move(9,px,py+c.pullTravel); events.clear(); c.cancel(); check(events.equals(List.of("5:0:0")));
            c.down(9,px,py,0); c.up(9,100); check(!events.contains("1:57:0")); // Corridor never nudges.
            events.clear(); c.down(9,nx,ny,0); c.move(9,nx,ny+50); c.up(9,100); check(events.isEmpty());
            c.down(9,nx,ny,0); c.up(9,501); check(events.isEmpty());
            c.panel=new InteractionGeometry.Rect(nx-10,ny-10,nx+10,ny+10);
            c.down(9,nx,ny,0); c.up(9,100); check(events.isEmpty()); c.panel=null;
            c.down(9,nx,ny,0); c.up(9,100); check(c.held[3]);
            c.key(62,true,0); releases.remove(0).run(); check(c.held[3]); c.key(62,false,0);
            c.down(9,nx,ny,0); c.up(9,100); c.clear(); events.clear(); releases.remove(0).run(); check(events.isEmpty());
            check(c.hit(0,0)==Controls.NONE);
        }
        events.clear(); c.key(66,true,0); c.key(66,true,1); c.key(66,true,0); c.key(66,false,0);
        check(events.equals(List.of("2:0:0","1:28:0")));
        c.key(20,true,0); check(c.held[2]); c.clear(); c.enabled=false; c.key(20,false,0); check(!c.held[2]);
        System.out.println("PASS: pointer ownership, immediate independent flippers, guard misses, deliberate single nudge, absolute plunger, cancellation and keyboard parity");
    }
}
