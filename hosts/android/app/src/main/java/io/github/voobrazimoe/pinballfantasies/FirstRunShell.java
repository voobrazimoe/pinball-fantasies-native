package io.github.voobrazimoe.pinballfantasies;

import android.animation.ValueAnimator;
import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Color;
import android.graphics.Paint;
import android.os.SystemClock;
import android.view.Gravity;
import android.view.View;
import android.widget.*;
import java.util.List;

/** Opaque no-data UI above the native surface. No engine controls live here. */
final class FirstRunShell extends FrameLayout {
    final Button button;
    /** Present only when the APK carries the official 10-minute demo. */
    final Button demoButton;
    boolean startingDemo;
    FirstRunShell(Context context, Runnable request) { this(context, request, null); }
    FirstRunShell(Context context, Runnable request, Runnable demo) {
        super(context);
        setBackgroundColor(Color.BLACK);
        setClickable(true);
        LinearLayout content = new LinearLayout(context);
        content.setOrientation(LinearLayout.VERTICAL);
        content.setGravity(Gravity.CENTER_VERTICAL);
        int pad = dp(24);
        content.setPadding(pad,pad,pad,pad);
        content.addView(new Label(context,"Pinball",Vga.LIGHT,2,false),lines(0));
        content.addView(new Label(context,"Fantasies",Vga.RED,2,false),lines(dp(6)));
        View rule = new View(context); rule.setBackgroundColor(Vga.SHADE);
        LinearLayout.LayoutParams ruled = new LinearLayout.LayoutParams(-1,dp(4)); ruled.topMargin=dp(12);
        content.addView(rule,ruled);
        content.addView(new Label(context,"C:> Import",Vga.MID,1,true),lines(dp(28)));
        content.addView(new Label(context,demo == null ? "Select the folder containing your original DOS game files."
                : "Import the full game from your original DOS folder, or play the official 10-minute Party Land demo.",
                Vga.FACE,1,false),lines(dp(20)));
        button = action(content,"Import DOS folder",VgaButton.PANEL,"1");
        button.setOnClickListener(v -> request.run());
        if (demo != null) {
            demoButton = action(content,"Play 10-minute demo",VgaButton.PRIMARY,"2");
            demoButton.setOnClickListener(v -> demo.run());
        } else demoButton = null;
        content.addView(new Label(context,"The files are validated and copied to private app storage.",
                Vga.MID,1,false),lines(dp(28)));
        ScrollView scroll = new ScrollView(context);
        scroll.setFillViewport(true);
        scroll.addView(content,new FrameLayout.LayoutParams(-1,-2));
        addView(scroll,new FrameLayout.LayoutParams(-1,-1));
    }
    private int dp(int value) { return Math.round(value*getResources().getDisplayMetrics().density); }
    private LinearLayout.LayoutParams lines(int top) {
        LinearLayout.LayoutParams lp = new LinearLayout.LayoutParams(-1,-2); lp.topMargin=top; return lp;
    }
    private Button action(LinearLayout parent,String label,int style,String key) {
        VgaButton b = new VgaButton(getContext(),label,style); b.hint=key;
        LinearLayout.LayoutParams lp = new LinearLayout.LayoutParams(-1,dp(60)); lp.topMargin=dp(18);
        parent.addView(b,lp);
        return b;
    }
    void present(boolean loaded, boolean busy, boolean keyboard, View menu, View touch) {
        setVisibility(loaded ? GONE : VISIBLE);
        menu.setVisibility(loaded && !keyboard ? VISIBLE : GONE);
        touch.setVisibility(loaded && !keyboard ? VISIBLE : GONE);
        button.setEnabled(!busy && !loaded);
        button.setText(busy && !startingDemo ? "Importing…" : "Import DOS folder");
        if (demoButton != null) {
            demoButton.setEnabled(!busy && !loaded);
            demoButton.setText(busy && startingDemo ? "Starting demo…" : "Play 10-minute demo");
        }
    }

    /** Wrapped pixel-font text at a whole multiple of the 16 dp cell; a prompt adds a blinking cursor. */
    static final class Label extends View {
        private static final long BLINK_MS=530;
        private final Paint paint=new Paint();
        private final String text;
        private final int color, size;
        private final boolean prompt;
        Label(Context context,String text,int color,int size,boolean prompt) {
            super(context); this.text=PixelFont.normalize(text); this.color=color; this.size=size; this.prompt=prompt;
            setContentDescription(text); setFocusable(false);
        }
        private int scale() { return size*PixelFont.scale(16*getResources().getDisplayMetrics().density); }
        private List<String> lines(int width) {
            int s=scale(); return PixelFont.wrap(text,width/(PixelFont.CELL*s)-(prompt?1:0));
        }
        @Override protected void onMeasure(int w,int h) {
            int width=MeasureSpec.getSize(w), s=scale();
            int count=lines(Math.max(1,width-getPaddingLeft()-getPaddingRight())).size();
            setMeasuredDimension(width,count*PixelFont.CELL*s+(count-1)*4*s);
        }
        @Override protected void onDraw(Canvas canvas) {
            int s=scale(); float y=0;
            List<String> rows=lines(getWidth());
            paint.setColor(color);
            for(String row:rows) { PixelFont.draw(canvas,row,0,y,s,paint); y+=PixelFont.CELL*s+4*s; }
            if(!prompt) return;
            long now=SystemClock.uptimeMillis();
            boolean animated=ValueAnimator.areAnimatorsEnabled();
            if(!animated || (now/BLINK_MS)%2==0) {
                paint.setColor(Vga.FACE);
                String last=rows.get(rows.size()-1);
                PixelFont.draw(canvas,"_",PixelFont.width(last,s),y-PixelFont.CELL*s-4*s,s,paint);
            }
            if(animated) postInvalidateDelayed(BLINK_MS-now%BLINK_MS);
        }
    }
}
