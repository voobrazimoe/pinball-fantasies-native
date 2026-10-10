package io.github.voobrazimoe.pinballfantasies;
public final class GeometryTest {
    static void check(boolean b) { if(!b) throw new AssertionError(); }
    public static void main(String[] args) {
        // Actual renderer snapshots: portrait full-table, landscape scrolling and 4:3 frontend.
        float[][] displays={{1080,2400,0,172,1080,2055,3},{2400,1080,480,0,1440,1080,3},
                {2400,1080,916,0,567,1080,3},{1920,1080,240,0,1440,1080,3},{1080,1920,0,0,1080,1920,3}};
        for(float[] v:displays) {
            float w=v[0],h=v[1],d=v[6];
            InteractionGeometry.Rect vp=new InteractionGeometry.Rect(v[2],v[3],v[2]+v[4],v[3]+v[5]);
            InteractionGeometry g=new InteractionGeometry(w,h,24,48,36,48,d,vp);
            check(g.left.width()>=100*d && g.left.height()>=80*d);
            check(g.right.width()==g.left.width() && !g.left.overlaps(g.right));
            check(g.left.left==g.safe.left && g.right.right==g.safe.right && g.left.bottom==g.safe.bottom);
            check(g.frame.left==Math.max(vp.left,g.safe.left));
            check(g.frame.right==Math.min(vp.right,g.safe.right));
            check(!g.nudge.overlaps(g.leftGuard) && !g.nudge.overlaps(g.rightGuard));
            check(!g.nudge.overlaps(g.plunger));
            int[] snapshot={(int)v[2],(int)v[3],(int)v[4],(int)v[5],(int)w,(int)h,320,w<h?609:240};
            InteractionGeometry.Rect actual=InteractionGeometry.viewport(snapshot,w,h);
            InteractionGeometry pull=new InteractionGeometry(w,h,24,48,36,48,d,actual,snapshot[7]);
            check(pull.plunger.top>=pull.frame.top+pull.frame.height()*33/snapshot[7]);
            check(!pull.plunger.overlaps(pull.leftGuard) && !pull.plunger.overlaps(pull.rightGuard));
            check(pull.plunger.width()>0 && pull.plunger.height()>0);
            check(actual.left==vp.left && actual.bottom==vp.bottom);
            InteractionGeometry.Rect menu=InteractionGeometry.mobileMenu(g.safe,actual,snapshot[7],d);
            check(menu.top>=actual.top+actual.height()*33/snapshot[7]);
            check(menu.left>=g.safe.left && menu.right<=g.safe.right && menu.bottom<=g.safe.bottom);
            check(menu.width()==48*d && menu.right<=actual.right);
            // Selector, Options, Players/Play all preserve upper original content.
            for(float panelHeight:new float[]{(w>h?152:224)*d,120*d,120*d}) {
                float y=InteractionGeometry.lowerTop(g.safe,actual,panelHeight,d);
                check(y>=g.safe.top && y+panelHeight<=g.safe.bottom);
                check(y+panelHeight>=Math.min(g.safe.bottom-8*d,actual.bottom));
                check(y>actual.top+actual.height()*.35f);
            }
            for(boolean selector:new boolean[]{true,false}) {
                InteractionGeometry.Panel p=new InteractionGeometry.Panel(w,h,24,48,36,48,d,selector,selector?5:4);
                check(p.fits(g.safe)); check(p.bounds.width()<=(selector && w>h?600:360)*d);
                for(InteractionGeometry.Rect cell:p.cells) {
                    check(cell.width()==p.cells[0].width() && cell.height()==p.cells[0].height());
                    check(cell.left>=p.bounds.left && cell.bottom<=p.bounds.bottom);
                }
                if(!selector) {
                    int y=p.topWithin(g.safe.top,Math.min(g.leftGuard.top,g.rightGuard.top));
                    InteractionGeometry.Rect utility=new InteractionGeometry.Rect(p.bounds.left,y,p.bounds.right,y+p.bounds.height());
                    check(!utility.overlaps(g.leftGuard) && !utility.overlaps(g.rightGuard));
                }
                if(w<h && selector) check(p.cells[4].left==p.cells[0].left && p.cells[4].right==p.cells[0].right);
            }
        }
        // A full bottom letterbox puts all contextual content outside the frame.
        InteractionGeometry.Rect safe=new InteractionGeometry.Rect(12,24,988,1876);
        InteractionGeometry.Rect frame=new InteractionGeometry.Rect(0,200,1000,1400);
        check(InteractionGeometry.lowerTop(safe,frame,224,1)==1400);
        System.out.println("PASS: responsive equal selector cells, capped panels, safe bounds, native viewport and guarded touch rectangles");
    }
}
