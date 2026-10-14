package io.github.voobrazimoe.pinballfantasies;

import android.animation.ValueAnimator;
import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.os.SystemClock;
import android.view.Gravity;
import java.util.List;

/** A Button that keeps its semantic text but draws itself as a raised VGA key in the pixel font. */
final class VgaButton extends android.widget.Button {
    static final int PANEL=0, PRIMARY=1, TAB=2;
    private static final long PULSE_MS=600;
    private final int style;
    private final Paint paint=new Paint();
    /** Optional key chip, e.g. "P" or "ESC", drawn before the label. */
    String hint;
    /** Text drawn instead of the semantic label, e.g. a glyph for an icon-only key. */
    String glyph;
    boolean pulse;

    VgaButton(Context context,String label,int style) {
        super(context);
        this.style=style;
        setText(label);
        setFocusable(false);
        setAllCaps(false);
        setBackground(null);
        setStateListAnimator(null);
        setMinWidth(0); setMinimumWidth(0); setMinHeight(0); setMinimumHeight(0);
        setPadding(0,0,0,0);
        setGravity(Gravity.CENTER);
        setTextColor(0);
    }

    @Override protected void drawableStateChanged() { super.drawableStateChanged(); invalidate(); }

    @Override protected void onDraw(Canvas canvas) {
        float d=getResources().getDisplayMetrics().density;
        int u=Vga.unit(d), w=getWidth(), h=getHeight();
        boolean pressed=isPressed(), enabled=isEnabled();
        int face=Vga.FACE, light=Vga.LIGHT, shade=Vga.SHADE, ink=Vga.BLACK, chip=Vga.SHADE;
        if(style==PRIMARY) { face=Vga.RED; light=Vga.RED_LIGHT; shade=Vga.RED_SHADE; ink=Vga.LIGHT; chip=Vga.RED_SHADE; }
        if(!enabled) { face=Vga.MID; light=Vga.FACE; ink=Vga.SHADE; chip=Vga.MID; }
        float l=0,t=0,r=w,b=h;
        if(style==TAB) {
            float tw=Math.min(w,40*d), th=Math.min(h,26*d);
            l=Math.round((w-tw)/2); r=l+Math.round(tw); b=Math.round(th);
        }
        if(pulse && enabled && ValueAnimator.areAnimatorsEnabled()) {
            long now=SystemClock.uptimeMillis();
            if((now/PULSE_MS)%2==0) {
                paint.setStyle(Paint.Style.FILL); paint.setColor(Vga.RED);
                canvas.drawRect(l,t,r,b,paint);
            }
            l+=2*u; t+=2*u; r-=2*u; b-=2*u;
            postInvalidateDelayed(PULSE_MS-now%PULSE_MS);
        }
        Vga.bevel(canvas,paint,l,t,r,b,u,face,light,shade,pressed);
        float shift=pressed?u:0;
        float pad=3*u+Math.max(u,4*d);
        float left=l+pad+shift, right=r-pad+shift, top=t+3*u+shift, bottom=b-3*u+shift;
        String label=PixelFont.normalize(glyph!=null ? glyph : getText().toString());
        String key=hint==null ? null : PixelFont.normalize(hint);
        int scale=PixelFont.scale(Math.min(16*d,(bottom-top)*.5f));
        if(key!=null) {
            int ks=scale;
            float kw=PixelFont.width(key,ks)+4*ks, kh=PixelFont.CELL*ks+4*ks;
            float ky=(top+bottom-kh)/2;
            paint.setColor(chip); canvas.drawRect(left,ky,left+kw,ky+kh,paint);
            paint.setColor(enabled?Vga.LIGHT:Vga.FACE);
            PixelFont.draw(canvas,key,left+2*ks,ky+2*ks,ks,paint);
            left+=kw+3*ks;
        }
        List<String> lines;
        while(true) {
            int columns=(int)Math.floor((right-left)/(PixelFont.CELL*scale));
            lines=PixelFont.wrap(label,columns);
            float textHeight=lines.size()*PixelFont.CELL*scale+(lines.size()-1)*3*scale;
            boolean fits=columns>0 && textHeight<=bottom-top && longestWord(label)<=columns;
            if(fits || scale==1) break;
            scale--;
        }
        float lineHeight=PixelFont.CELL*scale+3*scale;
        float y=(top+bottom-(lines.size()*lineHeight-3*scale))/2;
        paint.setColor(ink);
        for(String line:lines) {
            float x=key==null ? (left+right-PixelFont.width(line,scale))/2 : left;
            PixelFont.draw(canvas,line,Math.round(x),Math.round(y),scale,paint);
            y+=lineHeight;
        }
    }
    private static int longestWord(String text) {
        int n=0; for(String word:text.split(" ")) n=Math.max(n,word.length()); return n;
    }
}
