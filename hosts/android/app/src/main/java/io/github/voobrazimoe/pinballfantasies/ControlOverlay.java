package io.github.voobrazimoe.pinballfantasies;

import android.animation.ValueAnimator;
import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.os.SystemClock;
import android.view.MotionEvent;
import android.view.View;
import android.view.ViewConfiguration;
import android.view.animation.OvershootInterpolator;

final class ControlOverlay extends View {
    private final Controls controls;
    SemanticUi state;
    private final Paint paint = new Paint();
    private final int[] viewport=new int[8];
    void viewport(int[] bounds) {
        if(java.util.Arrays.equals(viewport,bounds)) return;
        System.arraycopy(bounds,0,viewport,0,8); updateGeometry();
    }
    private int safeLeft, safeTop, safeRight, safeBottom;
    ControlOverlay(Context context, Controls controls) {
        super(context); this.controls=controls;
        controls.changed=this::invalidate;
        setContentDescription("Game controls: outlined lower corners flippers; upper central playfield tap nudge; Pull corridor downward drag plunger");
    }
    void safeArea(int left, int top, int right, int bottom) {
        safeLeft=left; safeTop=top; safeRight=right; safeBottom=bottom;
        updateGeometry();
    }
    private void updateGeometry() {
        controls.geometry(getWidth(),getHeight(),safeLeft,safeTop,safeRight,safeBottom,
                ViewConfiguration.get(getContext()).getScaledTouchSlop(),ViewConfiguration.getLongPressTimeout(),
                getResources().getDisplayMetrics().density,frameBounds());
        invalidate();
    }
    private InteractionGeometry.Rect frameBounds() {
        return InteractionGeometry.viewport(viewport,getWidth(),getHeight());
    }
    @Override protected void onSizeChanged(int w,int h,int oldw,int oldh) {
        if (oldw>0 && oldh>0) controls.cancel();
        updateGeometry();
    }
    // Released spring: the drawn knob snaps back past its start, independent of engine input.
    private ValueAnimator spring;
    private float springX, springY, springOffset;
    private boolean wasPulling;
    private static final long BOB_MS=150, BLINK_MS=530;
    private static boolean animated() { return ValueAnimator.areAnimatorsEnabled(); }
    @Override protected void onDraw(Canvas canvas) {
        float d=getResources().getDisplayMetrics().density;
        int u=Vga.unit(d), s=PixelFont.scale(16*d);
        long now=SystemClock.uptimeMillis();
        paint.setStyle(Paint.Style.FILL);
        if (state!=null && state.mode==SemanticUi.STARTUP) {
            String text=PixelFont.normalize("Tap to continue");
            if (!animated() || (now/BLINK_MS)%2==0) {
                float x=Math.round((getWidth()-PixelFont.width(text,s))/2f), y=Math.round(getHeight()*.8f);
                paint.setColor(Vga.BLACK); PixelFont.draw(canvas,text,x+s,y+s,s,paint);
                paint.setColor(Vga.AMBER); PixelFont.draw(canvas,text,x,y,s,paint);
            }
            if (animated()) postInvalidateDelayed(BLINK_MS-now%BLINK_MS);
        }
        if (!controls.gameplay || controls.layout==null) return;
        for(int control:new int[]{Controls.LEFT,Controls.RIGHT}) {
            InteractionGeometry.Rect r=controls.region(control);
            boolean on=controls.held[control];
            if (on) { paint.setColor(0x38f3b245); canvas.drawRect(r.left,r.top,r.right,r.bottom,paint); }
            paint.setColor(on ? Vga.AMBER : 0x99f3b245);
            float inset=10*d, arm=Math.min(22*d,r.width()/4), line=Math.max(2*u,Math.round(3*d));
            canvas.drawRect(r.left+inset,r.top+inset,r.left+inset+arm,r.top+inset+line,paint);
            canvas.drawRect(r.left+inset,r.top+inset,r.left+inset+line,r.top+inset+arm,paint);
            canvas.drawRect(r.right-inset-arm,r.top+inset,r.right-inset,r.top+inset+line,paint);
            canvas.drawRect(r.right-inset-line,r.top+inset,r.right-inset,r.top+inset+arm,paint);
            String letter=control==Controls.LEFT?"L":"R";
            int big=2*s;
            PixelFont.draw(canvas,letter,Math.round(r.cx()-PixelFont.CELL*big/2f),Math.round(r.cy()-PixelFont.CELL*big/2f),big,paint);
        }
        if (!controls.plungerAvailable) { wasPulling=false; return; }
        InteractionGeometry.Rect c=controls.layout.plunger;
        Controls.Pointer p=controls.spring();
        boolean springing=spring!=null && spring.isRunning();
        if (p!=null) {
            springX=p.x; springY=p.y; springOffset=p.position/32f*p.travel; wasPulling=true;
            if (springing) spring.cancel();
            paint.setColor(0x8cf3b245);
            canvas.drawRect(c.left,c.top,c.left+u,c.bottom,paint);
        } else if (wasPulling) {
            wasPulling=false;
            spring=ValueAnimator.ofFloat(springOffset,0f);
            spring.setDuration(260); spring.setInterpolator(new OvershootInterpolator(3f));
            spring.addUpdateListener(a->{ springOffset=(float)a.getAnimatedValue(); invalidate(); });
            spring.start(); springing=true;
        }
        if (p==null && !springing) {
            String pull=PixelFont.normalize("Pull");
            float y=Math.round(c.top+c.height()*.62f), x=Math.round(c.cx()-PixelFont.width(pull,s)/2f);
            paint.setColor(0xb8000000);
            canvas.drawRect(x-3*s,y-3*s,x+PixelFont.width(pull,s)+3*s,y+PixelFont.CELL*s+2*s,paint);
            paint.setColor(Vga.AMBER);
            canvas.drawRect(x-3*s,y-3*s,x+PixelFont.width(pull,s)+3*s,y-2*s,paint);
            canvas.drawRect(x-3*s,y+PixelFont.CELL*s+s,x+PixelFont.width(pull,s)+3*s,y+PixelFont.CELL*s+2*s,paint);
            canvas.drawRect(x-3*s,y-3*s,x-2*s,y+PixelFont.CELL*s+2*s,paint);
            canvas.drawRect(x+PixelFont.width(pull,s)+2*s,y-3*s,x+PixelFont.width(pull,s)+3*s,y+PixelFont.CELL*s+2*s,paint);
            PixelFont.draw(canvas,pull,x,y,s,paint);
            int[] bob={0,1,2,1};
            float shift=animated() ? bob[(int)((now/BOB_MS)%4)]*s : 0;
            String arrow=String.valueOf(PixelFont.LOWER);
            float ax=Math.round(c.cx()-PixelFont.CELL*s/2f);
            PixelFont.draw(canvas,arrow,ax,y+12*s+shift,s,paint);
            paint.setColor(0x8cf3b245);
            PixelFont.draw(canvas,arrow,ax,y+20*s+shift,s,paint);
            if (animated()) postInvalidateDelayed(BOB_MS-now%BOB_MS);
        }
        if (p==null && !springing) return;
        // Charge meter: eight segments filling upward with the drawn spring travel.
        int lit=p==null ? 0 : Math.round(p.position/32f*8);
        float segW=7*u*2, segH=Math.round(4.5f*d), gap=Math.round(1.5f*d);
        float mx=Math.round(c.left+6*d), mBottom=Math.round(c.bottom-8*d), mTop=mBottom-8*segH-7*gap;
        paint.setColor(0xb8000000);
        canvas.drawRect(mx-2*u,mTop-2*u,mx+segW+2*u,mBottom+2*u,paint);
        for (int i=0;i<8;i++) {
            boolean on=8-i<=lit;
            paint.setColor(on ? (i<2 ? Vga.AMBER_HOT : Vga.AMBER) : Vga.DOT_OFF);
            float top=mTop+i*(segH+gap);
            canvas.drawRect(mx,top,mx+segW,top+segH,paint);
        }
        // The finger's rubber band: a rail from the touch-down point to a knob at the charge.
        float knob=springY+springOffset;
        paint.setColor(0xb3f3b245);
        canvas.drawRect(springX-u,Math.min(springY,knob),springX+u,Math.max(springY,knob),paint);
        float kw=36*d, kh=20*d;
        Vga.bevel(canvas,paint,Math.round(springX-kw/2),Math.round(knob-kh/2),Math.round(springX+kw/2),Math.round(knob+kh/2),
                u,Vga.FACE,Vga.LIGHT,Vga.SHADE,false);
    }
    @Override public boolean onTouchEvent(MotionEvent event) {
        if (controls.keyboardMode) return false;
        if (!controls.gameplay) {
            if (state!=null && state.mode==SemanticUi.STARTUP && event.getActionMasked()==MotionEvent.ACTION_UP)
                controls.tap(57);
            return true;
        }
        int index=event.getActionIndex(), id=event.getPointerId(index);
        switch(event.getActionMasked()) {
            case MotionEvent.ACTION_DOWN: case MotionEvent.ACTION_POINTER_DOWN:
                controls.down(id,event.getX(index),event.getY(index),event.getEventTime()); break;
            case MotionEvent.ACTION_MOVE:
                for(int i=0;i<event.getPointerCount();i++)
                    controls.move(event.getPointerId(i),event.getX(i),event.getY(i)); break;
            case MotionEvent.ACTION_UP: case MotionEvent.ACTION_POINTER_UP:
                controls.move(id,event.getX(index),event.getY(index)); controls.up(id,event.getEventTime()); break;
            case MotionEvent.ACTION_CANCEL: controls.cancel(); break;
            default: break;
        }
        return true;
    }
}
