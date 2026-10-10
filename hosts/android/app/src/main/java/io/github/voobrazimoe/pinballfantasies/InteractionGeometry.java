package io.github.voobrazimoe.pinballfantasies;

// Host-only pixel geometry. Rectangles are also the overlay's drawing primitives.
final class InteractionGeometry {
    static final class Rect {
        final float left,top,right,bottom;
        Rect(float l,float t,float r,float b) { left=l; top=t; right=Math.max(l,r); bottom=Math.max(t,b); }
        float width() { return right-left; } float height() { return bottom-top; }
        float cx() { return (left+right)/2; } float cy() { return (top+bottom)/2; }
        boolean contains(float x,float y) { return x>=left && x<right && y>=top && y<bottom; }
        boolean overlaps(Rect r) { return left<r.right && right>r.left && top<r.bottom && bottom>r.top; }
    }
    final Rect safe,frame,left,right,leftGuard,rightGuard,nudge,plunger;
    InteractionGeometry(float w,float h,float l,float t,float r,float b,float density,Rect viewport) {
        this(w,h,l,t,r,b,density,viewport,383);
    }
    InteractionGeometry(float w,float h,float l,float t,float r,float b,float density,Rect viewport,int sourceHeight) {
        safe=new Rect(l,t,w-r,h-b);
        frame=new Rect(Math.max(l,viewport.left),Math.max(t,viewport.top),
                Math.min(w-r,viewport.right),Math.min(h-b,viewport.bottom));
        float gap=8*density, guard=24*density;
        float zoneHeight=Math.min(safe.height()*.30f,160*density);
        float zoneWidth=Math.min(safe.width()*.46f,200*density);
        // On wide displays extend into side letterboxes, while keeping a thumb-sized cap.
        if(w>h) zoneWidth=Math.min(safe.width()*.40f,
                Math.max(160*density,Math.max(frame.left-l,w-r-frame.right)+48*density));
        zoneWidth=Math.min(zoneWidth,240*density);
        left=new Rect(l,safe.bottom-zoneHeight,l+zoneWidth,safe.bottom);
        right=new Rect(safe.right-zoneWidth,safe.bottom-zoneHeight,safe.right,safe.bottom);
        leftGuard=new Rect(l,Math.max(t,left.top-guard),Math.min(safe.right,left.right+guard),safe.bottom);
        rightGuard=new Rect(Math.max(l,right.left-guard),Math.max(t,right.top-guard),safe.right,safe.bottom);
        float neutralBottom=Math.min(frame.bottom,Math.min(leftGuard.top,rightGuard.top)-gap);
        float corridorWidth=Math.min(frame.width()*.25f,96*density);
        float pullTop=frame.top+frame.height()*33/Math.max(33,sourceHeight)+gap;
        plunger=new Rect(frame.right-corridorWidth,pullTop,frame.right,neutralBottom);
        nudge=new Rect(frame.left+gap,frame.top+gap,plunger.left-gap,neutralBottom);
    }
    // Renderer coordinates remain authoritative; no Java aspect-ratio policy.
    static Rect viewport(int[] v,float w,float h) {
        if(v[4]<=0 || v[5]<=0 || (v[4]>v[5])!=(w>h)) return new Rect(0,0,0,0);
        float sx=w/v[4],sy=h/v[5];
        return new Rect(v[0]*sx,v[1]*sy,(v[0]+v[2])*sx,(v[1]+v[3])*sy);
    }
    static float lowerTop(Rect safe,Rect frame,float height,float d) {
        // Use bottom letterbox when it can hold the panel; otherwise straddle
        // the framebuffer bottom with a small gap. Never centre over the table.
        float end=frame.height()>0 ? Math.min(safe.bottom,frame.bottom+height+8*d) : safe.bottom;
        return Math.max(safe.top,end-height-8*d);
    }
    static Rect mobileMenu(Rect safe,Rect frame,int sourceHeight,float d) {
        float size=Math.min(48*d,Math.min(safe.width(),safe.height()));
        float matrixBottom=frame.top+frame.height()*33/Math.max(33,sourceHeight);
        float x=Math.max(safe.left,Math.min(frame.right-size-4*d,safe.right-size));
        float y=Math.max(safe.top,Math.min(matrixBottom+4*d,safe.bottom-size));
        return new Rect(x,y,x+size,y+size);
    }
    static final class Panel {
        final Rect bounds; final Rect[] cells;
        final int columns,rows;
        Panel(float w,float h,float l,float t,float r,float b,float d,boolean selector,int count) {
            columns=selector ? (w>h?3:2) : Math.min(2,Math.max(1,count));
            rows=(count+columns-1)/columns;
            float gap=8*d, padding=8*d;
            float width=Math.min(w-l-r-24*d,(selector && w>h ? 600:360)*d);
            float cellHeight=(selector?64:48)*d;
            float height=rows*cellHeight+(rows-1)*gap+2*padding;
            float x=l+(w-l-r-width)/2, y=t+(h-t-b-height)/2;
            bounds=new Rect(x,y,x+width,y+height); cells=new Rect[count];
            float cw=(width-2*padding-(columns-1)*gap)/columns;
            for(int i=0;i<count;i++) {
                float cx=x+padding+(i%columns)*(cw+gap),cy=y+padding+(i/columns)*(cellHeight+gap);
                cells[i]=new Rect(cx,cy,cx+cw,cy+cellHeight);
            }
        }
        int topWithin(float safeTop,float availableBottom) {
            return Math.round(safeTop+Math.max(0,(availableBottom-safeTop-bounds.height())/2));
        }
        boolean fits(Rect safe) { return bounds.left>=safe.left && bounds.top>=safe.top && bounds.right<=safe.right && bounds.bottom<=safe.bottom; }
    }
}
