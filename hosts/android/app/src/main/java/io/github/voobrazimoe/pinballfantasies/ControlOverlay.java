package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.view.MotionEvent;
import android.view.View;
import android.view.ViewConfiguration;

final class ControlOverlay extends View {
    private final Controls controls;
    SemanticUi state;
    private final Paint paint = new Paint(Paint.ANTI_ALIAS_FLAG);
    private final int[] viewport=new int[8];
    void viewport(int[] bounds) {
        if(java.util.Arrays.equals(viewport,bounds)) return;
        System.arraycopy(bounds,0,viewport,0,8); updateGeometry();
    }
    private int safeLeft, safeTop, safeRight, safeBottom;
    ControlOverlay(Context context, Controls controls) {
        super(context); this.controls=controls;
        controls.changed=this::invalidate;
        setContentDescription("Game controls: lower corners flippers; central playfield tap nudge; blue rectangle: drag down, then release to launch the ball");
    }
    void safeArea(int left, int top, int right, int bottom) {
        safeLeft=left; safeTop=top; safeRight=right; safeBottom=bottom;
        updateGeometry();
    }
    private void updateGeometry() {
        controls.geometry(getWidth(),getHeight(),safeLeft,safeTop,safeRight,safeBottom,
                ViewConfiguration.get(getContext()).getScaledTouchSlop(),ViewConfiguration.getLongPressTimeout(),
                getResources().getDisplayMetrics().density,frameBounds(),viewport[7]);
        invalidate();
    }
    private InteractionGeometry.Rect frameBounds() {
        return InteractionGeometry.viewport(viewport,getWidth(),getHeight());
    }
    @Override protected void onSizeChanged(int w,int h,int oldw,int oldh) {
        if (oldw>0 && oldh>0) controls.cancel();
        updateGeometry();
    }
    @Override protected void onDraw(Canvas canvas) {
        paint.setColor(0x99ffffff); paint.setTextSize(14*getResources().getDisplayMetrics().scaledDensity);
        paint.setTextAlign(Paint.Align.CENTER);
        if (state!=null && state.mode==SemanticUi.STARTUP) {
            canvas.drawText("Tap to continue",getWidth()/2f,getHeight()*.8f,paint);
        }
        if (!controls.gameplay) return;
        for(int control:new int[]{Controls.LEFT,Controls.RIGHT}) {
            InteractionGeometry.Rect r=controls.region(control);
            paint.setStyle(Paint.Style.FILL); paint.setColor(0x18ffffff);
            canvas.drawRoundRect(r.left,r.top,r.right,r.bottom,12,12,paint);
            paint.setStyle(Paint.Style.STROKE); paint.setStrokeWidth(getResources().getDisplayMetrics().density);
            paint.setColor(0x66ffffff); canvas.drawRoundRect(r.left,r.top,r.right,r.bottom,12,12,paint);
            paint.setStyle(Paint.Style.FILL); paint.setColor(0x99ffffff);
            canvas.drawText(control==Controls.LEFT?"L":"R",r.cx(),r.cy()-(paint.ascent()+paint.descent())/2,paint);
        }
        if (controls.plungerAvailable) {
            InteractionGeometry.Rect r=controls.layout.plunger;
            float d=getResources().getDisplayMetrics().density;
            paint.setStyle(Paint.Style.FILL); paint.setColor(0x5541b6dc);
            canvas.drawRoundRect(r.left,r.top,r.right,r.bottom,8*d,8*d,paint);
            paint.setStyle(Paint.Style.STROKE); paint.setStrokeWidth(2*d);paint.setColor(0xbba3e4fa);
            canvas.drawRoundRect(r.left+d,r.top+d,r.right-d,r.bottom-d,8*d,8*d,paint);
            paint.setStyle(Paint.Style.FILL);paint.setColor(0xeeffffff);
            paint.setTextSize(Math.min(14*getResources().getDisplayMetrics().scaledDensity,r.width()/6));
            float centre=r.cy();
            canvas.drawText("DRAG DOWN",r.cx(),centre-16*d,paint);
            canvas.drawText("↓",r.cx(),centre+4*d,paint);
            paint.setTextSize(Math.min(11*getResources().getDisplayMetrics().scaledDensity,r.width()/10));
            canvas.drawText("RELEASE TO LAUNCH",r.cx(),centre+22*d,paint);
        }
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
