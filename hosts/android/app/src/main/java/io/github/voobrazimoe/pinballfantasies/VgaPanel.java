package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.widget.FrameLayout;

/** Raised frontend-coloured sheet with a dark title strip, like the selector's own panels. */
final class VgaPanel extends FrameLayout {
    private final Paint paint=new Paint();
    private String title="", status="";
    private int strip;
    private boolean framed=true;

    VgaPanel(Context context) {
        super(context);
        setWillNotDraw(false);
        setClipToPadding(true);
        title(null,null);
    }
    private int scale() { return PixelFont.scale(16*getResources().getDisplayMetrics().density); }

    /** Sets the strip text; a null title removes the strip. */
    void title(String title,String status) {
        int s=scale();
        this.title=title==null ? "" : PixelFont.normalize(title);
        this.status=status==null ? "" : PixelFont.normalize(status);
        strip=title==null ? 0 : PixelFont.CELL*s+6*s;
        framed(framed);
    }
    /** An unframed panel draws nothing, so its children stand on the game alone. */
    void framed(boolean framed) {
        int u=Vga.unit(getResources().getDisplayMetrics().density);
        this.framed=framed;
        if(framed) setPadding(3*u,3*u+strip,3*u,3*u); else setPadding(0,0,0,0);
        invalidate();
    }

    @Override protected void onDraw(Canvas canvas) {
        if(!framed) return;
        int u=Vga.unit(getResources().getDisplayMetrics().density), s=scale();
        Vga.bevel(canvas,paint,0,0,getWidth(),getHeight(),u,Vga.FACE,Vga.LIGHT,Vga.SHADE,false);
        if(strip==0) return;
        paint.setColor(Vga.SHADE);
        canvas.drawRect(3*u,3*u,getWidth()-3*u,3*u+strip,paint);
        // Long table names shrink to the strip rather than run under the status.
        int t=s;
        while(t>1 && PixelFont.width(title,t)+PixelFont.width(status,t)+(status.isEmpty()?8:12)*t>getWidth()-6*u) t--;
        float y=3*u+(strip-PixelFont.CELL*t)/2f;
        paint.setColor(Vga.FACE);
        PixelFont.draw(canvas,title,3*u+4*t,Math.round(y),t,paint);
        paint.setColor(Vga.AMBER);
        PixelFont.draw(canvas,status,getWidth()-3*u-4*t-PixelFont.width(status,t),Math.round(y),t,paint);
    }
}
