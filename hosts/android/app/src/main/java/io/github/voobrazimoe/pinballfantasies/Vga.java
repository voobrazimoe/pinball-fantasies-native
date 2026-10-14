package io.github.voobrazimoe.pinballfantasies;

import android.graphics.Canvas;
import android.graphics.Paint;

/** Overlay look: the selector frontend's own colours and raised VGA bevels. */
final class Vga {
    static final int FACE=0xffc3c3f3, LIGHT=0xfff3f3f3, MID=0xff717192, SHADE=0xff414161, BLACK=0xff000000,
            RED=0xffb24141, RED_LIGHT=0xfff37171, RED_SHADE=0xff611f1f,
            AMBER=0xfff3b245, AMBER_HOT=0xfff34141, DOT=0xff515151, DOT_OFF=0xff2a2a2a;
    private Vga() {}

    /** One logical pixel of the overlay: whole device pixels, two per dp step on dense screens. */
    static int unit(float density) { return Math.max(1,Math.round(density)); }

    /** Fills a raised (or sunken, when pressed) face with a one-unit black outline. */
    static void bevel(Canvas c,Paint p,float l,float t,float r,float b,int u,int face,int light,int shade,boolean pressed) {
        p.setStyle(Paint.Style.FILL);
        p.setColor(BLACK); c.drawRect(l,t,r,b,p);
        l+=u; t+=u; r-=u; b-=u;
        p.setColor(pressed?shade:light);
        c.drawRect(l,t,r,t+2*u,p); c.drawRect(l,t,l+2*u,b,p);
        p.setColor(pressed?light:shade);
        c.drawRect(l,b-2*u,r,b,p); c.drawRect(r-2*u,t,r,b,p);
        p.setColor(face); c.drawRect(l+2*u,t+2*u,r-2*u,b-2*u,p);
    }
}
