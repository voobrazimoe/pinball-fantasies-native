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
    private int safeLeft, safeTop, safeRight, safeBottom;
    ControlOverlay(Context context, Controls controls) {
        super(context); this.controls=controls;
        controls.changed=this::invalidate;
        setContentDescription("Game controls: bottom left/right flippers; table tap nudge; right table downward drag plunger");
    }
    void safeArea(int left, int top, int right, int bottom) {
        safeLeft=left; safeTop=top; safeRight=right; safeBottom=bottom;
        updateGeometry();
    }
    private void updateGeometry() {
        controls.geometry(getWidth(),getHeight(),safeLeft,safeTop,safeRight,safeBottom,
                ViewConfiguration.get(getContext()).getScaledTouchSlop(),ViewConfiguration.getLongPressTimeout());
        invalidate();
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
        float baseline=controls.labelY()-(paint.ascent()+paint.descent())/2;
        canvas.drawText("L",controls.labelX(Controls.LEFT),baseline,paint);
        canvas.drawText("R",controls.labelX(Controls.RIGHT),baseline,paint);
        if (controls.plungerAvailable) canvas.drawText("Pull ↓",controls.labelX(Controls.RIGHT),
                controls.top+(controls.stripTop-controls.top)*.5f,paint);
    }
    @Override public boolean onTouchEvent(MotionEvent event) {
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
