package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.view.MotionEvent;
import android.view.View;

final class ControlOverlay extends View {
    private final Controls controls;
    private final Paint paint = new Paint(Paint.ANTI_ALIAS_FLAG);
    ControlOverlay(Context context, Controls controls) {
        super(context); this.controls=controls;
        setContentDescription("Game controls: lower left/right flippers, lower right pull plunger, middle right nudge");
    }
    @Override protected void onDraw(Canvas canvas) {
        paint.setColor(0x22333333);
        float w=getWidth(), h=getHeight();
        canvas.drawRect(0,.65f*h,.4f*w,h,paint);
        canvas.drawRect(.4f*w,.65f*h,.8f*w,h,paint);
        canvas.drawRect(.8f*w,.65f*h,w,h,paint);
        paint.setColor(0x99ffffff); paint.setTextSize(14*getResources().getDisplayMetrics().scaledDensity);
        canvas.drawText("L",.05f*w,.95f*h,paint);
        canvas.drawText("R",.5f*w,.95f*h,paint);
        canvas.drawText("Pull ↓",.81f*w,.95f*h,paint);
        canvas.drawText("Nudge",.81f*w,.55f*h,paint);
    }
    @Override public boolean onTouchEvent(MotionEvent event) {
        int index=event.getActionIndex(), id=event.getPointerId(index);
        switch(event.getActionMasked()) {
            case MotionEvent.ACTION_DOWN: case MotionEvent.ACTION_POINTER_DOWN:
                controls.down(id,event.getX(index)/getWidth(),event.getY(index)/getHeight()); break;
            case MotionEvent.ACTION_MOVE:
                for(int i=0;i<event.getPointerCount();i++)
                    controls.move(event.getPointerId(i),event.getY(i)/getHeight()); break;
            case MotionEvent.ACTION_UP: case MotionEvent.ACTION_POINTER_UP:
                controls.move(id,event.getY(index)/getHeight()); controls.up(id); break;
            case MotionEvent.ACTION_CANCEL: controls.cancel(); break;
            default: break;
        }
        return true;
    }
}
