package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.view.Gravity;
import android.view.MotionEvent;
import android.view.View;
import android.view.animation.DecelerateInterpolator;
import android.view.animation.OvershootInterpolator;
import android.widget.Button;
import android.widget.FrameLayout;
import android.widget.LinearLayout;
import android.widget.ScrollView;

// In-Activity sheets never create a Window or change audio/lifecycle eligibility.
final class ControlMenu extends FrameLayout {
    final SemanticUi state=new SemanticUi();
    private final Controls controls;
    final Button menu;
    private final VgaButton tab;
    final VgaPanel sheet;
    /** Landscape Options only: Select and Back in a short strip under the original settings list. */
    final VgaPanel lower;
    private LinearLayout lowerContent;
    private LinearLayout content;
    private FrameLayout selector;
    private boolean scrollable;
    private boolean opened, advanced;
    private boolean pulled;
    private float pullStart;
    private String shown="";
    private boolean enter;
    private int safeLeft,safeTop,safeRight,safeBottom;
    private final Paint scrim=new Paint();
    private float dim;
    private android.animation.ValueAnimator fade;
    private boolean swallow;
    /** Landscape: the contextual panel stands as one column in the right letterbox, beside the original screen. */
    private boolean side;
    ControlMenu(Context context, Controls controls) {
        super(context); this.controls=controls;
        sheet=new VgaPanel(context);
        sheet.setOnTouchListener((v,e)->false);
        addView(sheet,new FrameLayout.LayoutParams(LayoutParams.MATCH_PARENT,LayoutParams.WRAP_CONTENT,Gravity.BOTTOM));
        lower=new VgaPanel(context); lower.setVisibility(GONE);
        addView(lower,new FrameLayout.LayoutParams(LayoutParams.WRAP_CONTENT,LayoutParams.WRAP_CONTENT,Gravity.TOP|Gravity.LEFT));
        tab=new VgaButton(context,"Menu",VgaButton.TAB); tab.glyph="▼";
        menu=tab; menu.setContentDescription("Mobile menu");
        addView(menu,new FrameLayout.LayoutParams(dp(44),dp(44),Gravity.TOP|Gravity.RIGHT));
        menu.setOnClickListener(v->{
            if(pulled) { pulled=false; return; }
            toggle();
        });
        // Pulling the tab down opens the menu as soon as the drag clearly starts.
        menu.setOnTouchListener((v,e)->{
            switch(e.getActionMasked()) {
                case MotionEvent.ACTION_DOWN: pulled=false; pullStart=e.getRawY(); break;
                case MotionEvent.ACTION_MOVE:
                    if(!pulled && !opened && !advanced && e.getRawY()-pullStart>dp(16)) { pulled=true; toggle(); }
                    break;
                default: break;
            }
            return false;
        });
        setClipChildren(true); rebuild();
    }
    private boolean modal() { return opened || advanced; }
    private static boolean hit(View v,MotionEvent e) {
        return v.getVisibility()==VISIBLE && e.getX()>=v.getLeft() && e.getX()<v.getRight() && e.getY()>=v.getTop() && e.getY()<v.getBottom();
    }
    // An open menu is modal: a touch anywhere outside it only closes it, the game never sees that gesture.
    @Override public boolean dispatchTouchEvent(MotionEvent e) {
        if(e.getActionMasked()==MotionEvent.ACTION_DOWN) {
            swallow=modal() && !hit(sheet,e) && !hit(menu,e);
            if(swallow) dismiss();
        }
        if(swallow) return true;
        return super.dispatchTouchEvent(e);
    }
    @Override protected void dispatchDraw(Canvas canvas) {
        if(dim>0) { scrim.setColor(Math.round(dim*0x99)<<24); canvas.drawRect(0,0,getWidth(),getHeight(),scrim); }
        super.dispatchDraw(canvas);
    }
    private void fadeScrim() {
        float target=modal() ? 1 : 0;
        if(fade!=null) fade.cancel();
        if(dim==target) return;
        fade=android.animation.ValueAnimator.ofFloat(dim,target);
        fade.setDuration(180);
        fade.addUpdateListener(a->{ dim=(float)a.getAnimatedValue(); invalidate(); });
        fade.start();
    }
    private void toggle() { controls.cancel(); opened=!opened; advanced=false; rebuild(); }
    private final int[] viewport=new int[8];
    void viewport(int[] bounds) {
        if(java.util.Arrays.equals(viewport,bounds)) return;
        System.arraycopy(bounds,0,viewport,0,8); place();
    }
    void keyboardMode(boolean present) {
        if(controls.keyboardMode==present) return;
        controls.keyboardMode(present);
        opened=advanced=false;
        setVisibility(present ? GONE : VISIBLE);
        rebuild();
    }
    private int dp(int value) { return Math.round(value*getResources().getDisplayMetrics().density); }
    boolean dismiss() {
        if (!opened && !advanced) return false;
        opened=advanced=false; rebuild(); return true;
    }
    void snapshot(int mode,int table,int flags) {
        if (!state.update(mode,table,flags)) return;
        boolean playing=state.gameplay();
        if (controls.gameplay!=playing || (controls.plungerAvailable && !state.plunger())) controls.cancel();
        controls.gameplay=playing; controls.plungerAvailable=state.plunger(); controls.changed.run();
        opened=advanced=false; rebuild();
    }
    private VgaButton button(LinearLayout row,String label,Runnable action) { return button(row,label,1,action); }
    // Every key keeps half a gap on each side, so equal weights give equal keys even beside spacers.
    private VgaButton button(LinearLayout row,String label,float weight,Runnable action) {
        VgaButton b=new VgaButton(getContext(),label,label.equals("Play")?VgaButton.PRIMARY:VgaButton.PANEL);
        b.hint=hint(label); b.pulse=label.equals("Play");
        LinearLayout.LayoutParams lp=new LinearLayout.LayoutParams(0,-1,weight);
        lp.leftMargin=lp.rightMargin=gap()/2;
        row.addView(b,lp);
        b.setOnClickListener(v->action.run()); return b;
    }
    private void spacer(LinearLayout row,float weight) { row.addView(new View(getContext()),new LinearLayout.LayoutParams(0,-1,weight)); }
    private int gap() { return dp(advanced ? 4 : opened ? 6 : 8); }
    private int keyHeight() { return dp(advanced ? 40 : opened ? 44 : 48); }
    /** The DOS key each touch action stands for, shown as a chip on its button. */
    static String hint(String label) {
        switch(label) {
            case "Pause": return "P";
            case "Advanced keyboard": return "F1-8";
            case "Music": return "M";
            case "Exit table": case "Back to tables": case "Back": return "ESC";
            case "Resume": case "Select": return "ENT";
            case "Up": return "↑";
            case "Down": return "↓";
            case "Yes": return "Y";
            case "No": return "N";
            default: return null;
        }
    }
    private LinearLayout row(LinearLayout parent) {
        LinearLayout row=new LinearLayout(getContext());
        LinearLayout.LayoutParams lp=new LinearLayout.LayoutParams(-1,keyHeight());
        if(parent.getChildCount()>0) lp.topMargin=gap();
        parent.addView(row,lp); return row;
    }
    private void action(String label) {
        controls.tap(state.code(label));
        if (opened) { opened=false; rebuild(); }
    }
    private String tableName() {
        return state.table>=1 && state.table<=4 ? SemanticUi.TABLES[state.table-1] : "Pinball";
    }
    private void title() {
        sheet.framed(selector==null);
        if(advanced) sheet.title("Keyboard",null);
        else if(opened) {
            if(state.mode==SemanticUi.PLAYING || state.mode==SemanticUi.ATTRACT) sheet.title(tableName(),null);
            else sheet.title("Menu",null);
        }
        else switch(state.mode) {
            case SemanticUi.ATTRACT: sheet.title(tableName(),null); break;
            case SemanticUi.OPTIONS: sheet.title("Options",null); break;
            case SemanticUi.PAUSED: sheet.title(tableName(),"Paused"); break;
            case SemanticUi.QUESTION: sheet.title(tableName(),"Sure?"); break;
            case SemanticUi.INITIALS: sheet.title("High score",null); break;
            default: sheet.title(null,null); break;
        }
    }
    private int sideRoom() {
        if(getWidth()<=getHeight()) return 0;
        InteractionGeometry.Rect frame=InteractionGeometry.viewport(viewport,getWidth(),getHeight());
        return frame.width()>0 ? Math.round(getWidth()-safeRight-frame.right) : 0;
    }
    private boolean wantSide() {
        boolean panel=state.mode==SemanticUi.ATTRACT || state.mode==SemanticUi.OPTIONS
                || state.mode==SemanticUi.PAUSED || state.mode==SemanticUi.QUESTION;
        return panel && !opened && !advanced && sideRoom()>=dp(120);
    }
    private void rebuild() {
        side=wantSide();
        sheet.removeAllViews();
        lower.removeAllViews(); lowerContent=null;
        selector=null; scrollable=advanced || (!opened && state.mode==SemanticUi.INITIALS);
        content=new LinearLayout(getContext()); content.setOrientation(LinearLayout.VERTICAL);
        content.setPadding(dp(8)-gap()/2,dp(8),dp(8)-gap()/2,dp(8));
        if(scrollable) { ScrollView scroll=new ScrollView(getContext()); scroll.addView(content); sheet.addView(scroll); }
        else sheet.addView(content);
        if (opened || advanced) {
            if (advanced) {
                // Rows of four keys; Space spans two and the tab or a tap outside closes the board.
                String[][] rows={{"Enter","Esc","Space"},{"P","M","Y","N"},{"↑","↓","←","→"}};
                int[][] codes={{28,1,57},{25,50,21,49},{72,80,75,77}};
                for(int i=0;i<rows.length;i++) {
                    LinearLayout r=row(content);
                    for(int j=0;j<rows[i].length;j++) {
                        final int code=codes[i][j];
                        button(r,rows[i][j],rows[i][j].equals("Space")?2:1,()->controls.tap(code));
                    }
                }
                for(int i=0;i<8;i+=4) { LinearLayout r=row(content); for(int j=i;j<i+4;j++) {
                    final int code=59+j; button(r,"F"+(j+1),()->controls.tap(code));
                }}
                letters(content);
            } else {
                // A short drop-down list under the tab, one full-width key per action.
                for(String label:state.utilities()) {
                    VgaButton b=button(row(content),label,()->{
                        if(label.equals("Advanced keyboard")) {advanced=true; opened=false; rebuild();}
                        else action(label);
                    });
                    if(label.equals("Advanced keyboard")) b.glyph="Keyboard →";
                }
            }
        } else if(state.mode==SemanticUi.SELECTOR || state.mode==SemanticUi.SELECTOR_TEXT) {
            sheet.removeAllViews(); selector=new FrameLayout(getContext()); sheet.addView(selector);
            for(int i=0;i<SemanticUi.TABLES.length;i++) {
                String label=SemanticUi.TABLES[i];
                VgaButton b=new VgaButton(getContext(),label,VgaButton.PANEL); b.hint="F"+(i+1);
                b.setOnClickListener(v->action(label)); selector.addView(b);
            }
        } else if(state.mode==SemanticUi.ATTRACT) {
            LinearLayout r=row(content);
            button(r,"−",()->{state.players(-1); rebuild();}).setContentDescription("Fewer players");
            Readout count=new Readout(getContext(),"Players "+state.players);
            LinearLayout.LayoutParams lp=new LinearLayout.LayoutParams(0,-1,2);
            lp.leftMargin=lp.rightMargin=gap()/2;
            r.addView(count,lp);
            button(r,"+",()->{state.players(1); rebuild();}).setContentDescription("More players");
            button(row(content),"Play",()->action("Play"));
        } else if(state.mode==SemanticUi.INITIALS) letters(content);
        else if(state.mode!=SemanticUi.STARTUP) {
            String[] actions=state.actions();
            int perRow=side ? 1 : 2;
            if(side && state.mode==SemanticUi.OPTIONS) {
                // ↑ ↓ stand beside the settings list; Select and Back sit in their own strip under it.
                actions=new String[]{"Up","Down"};
                lowerContent=new LinearLayout(getContext()); lowerContent.setOrientation(LinearLayout.VERTICAL);
                lowerContent.setPadding(dp(8)-gap()/2,dp(8),dp(8)-gap()/2,dp(8));
                lower.addView(lowerContent);
                LinearLayout r=row(lowerContent); r.getLayoutParams().height=dp(40);
                for(String label:new String[]{"Select","Back"}) button(r,label,()->action(label));
            }
            for(int i=0;i<actions.length;i+=perRow) {
                LinearLayout r=row(content);
                for(int j=i;j<Math.min(i+perRow,actions.length);j++) {String label=actions[j]; button(r,label,()->action(label));}
            }
        }
        title();
        tab.glyph=modal() ? "▲" : "▼"; tab.invalidate();
        fadeScrim();
        sheet.setVisibility(selector==null && content.getChildCount()==0 ? GONE : VISIBLE);
        lower.setVisibility(lowerContent==null ? GONE : VISIBLE);
        // The selector's text pages keep the same cards, so they do not replay the entrance.
        int mode=state.mode==SemanticUi.SELECTOR_TEXT ? SemanticUi.SELECTOR : state.mode;
        String key=mode+(opened?"o":"")+(advanced?"a":"");
        enter=sheet.getVisibility()==VISIBLE && !key.equals(shown);
        shown=sheet.getVisibility()==VISIBLE ? key : "";
        place();
    }
    private void letters(LinearLayout content) {
        // A ten-column grid: the shorter rows are centred with spacers so every letter key is the same size.
        for(String line:new String[]{"QWERTYUIOP","ASDFGHJKL","ZXCVBNM"}) {
            LinearLayout r=row(content);
            float side=(10-line.length())/2f;
            if(side>0) spacer(r,side);
            for(char letter:line.toCharArray()) {String label=String.valueOf(letter); button(r,label,()->action(label));}
            if(side>0) spacer(r,side);
        }
    }
    static androidx.core.graphics.Insets interactiveInsets(androidx.core.view.WindowInsetsCompat insets) {
        return insets.getInsets(androidx.core.view.WindowInsetsCompat.Type.systemBars()
                | androidx.core.view.WindowInsetsCompat.Type.displayCutout()
                | androidx.core.view.WindowInsetsCompat.Type.systemGestures());
    }
    void safeInsets(int left,int top,int right,int bottom) {
        safeLeft=left; safeTop=top; safeRight=right; safeBottom=bottom; place();
    }
    private void place() {
        if(getWidth()==0 || getHeight()==0) return;
        // Rotation or a new viewport can open or close the side letterbox; rebuild re-enters here once.
        if(wantSide()!=side) { rebuild(); return; }
        float d=getResources().getDisplayMetrics().density;
        FrameLayout.LayoutParams panel=(FrameLayout.LayoutParams)sheet.getLayoutParams();
        InteractionGeometry.Panel grid=new InteractionGeometry.Panel(getWidth(),getHeight(),safeLeft,safeTop,
                safeRight,safeBottom,d,selector!=null,selector!=null?5:Math.max(1,content.getChildCount()*2));
        int height;
        int padX=sheet.getPaddingLeft()+sheet.getPaddingRight(), padY=sheet.getPaddingTop()+sheet.getPaddingBottom();
        if(selector!=null) {
            height=Math.round(grid.bounds.height());
            for(int i=0;i<5;i++) {
                InteractionGeometry.Rect cell=grid.cells[i];
                FrameLayout.LayoutParams lp=new FrameLayout.LayoutParams(Math.round(cell.width()),Math.round(cell.height()));
                lp.leftMargin=Math.round(cell.left-grid.bounds.left); lp.topMargin=Math.round(cell.top-grid.bounds.top);
                selector.getChildAt(i).setLayoutParams(lp);
            }
        } else {
            content.measure(MeasureSpec.makeMeasureSpec(Math.round(grid.bounds.width())-padX,MeasureSpec.EXACTLY),
                    MeasureSpec.makeMeasureSpec(0,MeasureSpec.UNSPECIFIED));
            height=content.getMeasuredHeight()+padY;
            if(scrollable) height=Math.min(height,Math.max(0,getHeight()-safeTop-safeBottom-dp(64)));
        }
        InteractionGeometry.Rect frame=InteractionGeometry.viewport(viewport,getWidth(),getHeight());
        InteractionGeometry.Rect iconBounds=InteractionGeometry.mobileMenu(
                new InteractionGeometry.Rect(safeLeft,safeTop,getWidth()-safeRight,getHeight()-safeBottom),frame,viewport[7],d);
        int width=Math.round(grid.bounds.width());
        if(opened) {
            // The utility list is narrow; re-measure it at its own width.
            width=Math.min(width,dp(280));
            content.measure(MeasureSpec.makeMeasureSpec(width-padX,MeasureSpec.EXACTLY),MeasureSpec.makeMeasureSpec(0,MeasureSpec.UNSPECIFIED));
            height=content.getMeasuredHeight()+padY;
        }
        panel.gravity=Gravity.TOP|Gravity.LEFT; panel.width=width; panel.height=height;
        if(side) {
            width=Math.min(sideRoom()-dp(16),dp(200));
            content.measure(MeasureSpec.makeMeasureSpec(width-padX,MeasureSpec.EXACTLY),MeasureSpec.makeMeasureSpec(0,MeasureSpec.UNSPECIFIED));
            height=content.getMeasuredHeight()+padY;
            panel.width=width; panel.height=height;
        }
        int x=Math.round(grid.bounds.left);
        if(side) x=Math.round(frame.right+(getWidth()-safeRight-frame.right-width)/2f);
        if(opened) x=Math.round(Math.max(safeLeft,Math.min(iconBounds.right-width,getWidth()-safeRight-width)));
        int end=getHeight()-safeBottom;
        if(controls.gameplay && !scrollable && controls.layout!=null)
            end=Math.round(Math.min(controls.layout.leftGuard.top,controls.layout.rightGuard.top));
        InteractionGeometry.Rect safe=new InteractionGeometry.Rect(safeLeft,safeTop,getWidth()-safeRight,end);
        boolean contextual=!opened && !advanced && (state.mode==SemanticUi.SELECTOR || state.mode==SemanticUi.SELECTOR_TEXT
                || state.mode==SemanticUi.ATTRACT || state.mode==SemanticUi.OPTIONS || state.mode==SemanticUi.PAUSED
                || state.mode==SemanticUi.QUESTION);
        int y;
        if(side) y=safeTop+Math.max(0,(end-safeTop-height)/2);
        else if(contextual) y=Math.round(InteractionGeometry.lowerTop(safe,frame,height,d));
        else if(modal()) {
            // Menu and keyboard drop from just under the tab, lifted only as far as they must be to fit.
            if(advanced) height=panel.height=Math.min(height,Math.max(0,end-safeTop-dp(4)));
            y=Math.max(safeTop,Math.min(Math.round(iconBounds.bottom)+dp(4),end-height));
        }
        else y=scrollable ? safeTop+Math.max(0,(end-safeTop-height)/2) : grid.topWithin(safeTop,end);
        panel.setMargins(x,y,0,0); sheet.setLayoutParams(panel);
        if(lowerContent!=null) {
            // The original list (rows 14-174 of 240, x 164-400 of 640) leaves the frame's lowest quarter free.
            int lw=Math.round(Math.min(dp(280),frame.width()*.45f));
            lowerContent.measure(MeasureSpec.makeMeasureSpec(lw-padX,MeasureSpec.EXACTLY),MeasureSpec.makeMeasureSpec(0,MeasureSpec.UNSPECIFIED));
            int lh=lowerContent.getMeasuredHeight()+lower.getPaddingTop()+lower.getPaddingBottom();
            FrameLayout.LayoutParams lp=(FrameLayout.LayoutParams)lower.getLayoutParams();
            lp.gravity=Gravity.TOP|Gravity.LEFT; lp.width=lw; lp.height=lh;
            float bottom=Math.min(safe.bottom,frame.bottom)-dp(8);
            lp.setMargins(Math.round(frame.left+frame.width()*282/640f-lw/2f),Math.round(Math.max(safe.top,bottom-lh)),0,0);
            lower.setLayoutParams(lp);
        }
        controls.panel=!controls.keyboardMode && sheet.getVisibility()==VISIBLE ? new InteractionGeometry.Rect(x,y,x+panel.width,y+height) : null;
        FrameLayout.LayoutParams icon=(FrameLayout.LayoutParams)menu.getLayoutParams();
        icon.gravity=Gravity.TOP|Gravity.LEFT;
        icon.width=Math.round(iconBounds.width()); icon.height=Math.round(iconBounds.height());
        icon.setMargins(Math.round(iconBounds.left),Math.round(iconBounds.top),0,0); menu.setLayoutParams(icon);
        boolean visible=!controls.keyboardMode && frame.width()>0 && frame.height()>0;
        menu.setVisibility(visible ? VISIBLE : GONE);
        controls.menu=visible ? iconBounds : null;
        if(enter) { enter=false; reveal(contextual); }
    }
    /** Slides a newly shown sheet into place; layout bounds stay final, only its drawing moves. */
    private void reveal(boolean fromBelow) {
        sheet.animate().cancel();
        if(selector!=null) {
            sheet.setTranslationY(0); sheet.setAlpha(1);
            for(int i=0;i<selector.getChildCount();i++) {
                View cell=selector.getChildAt(i);
                cell.animate().cancel();
                cell.setTranslationY(dp(20)); cell.setAlpha(0);
                cell.animate().translationY(0).alpha(1).setStartDelay(45L*i).setDuration(220)
                        .setInterpolator(new DecelerateInterpolator(1.6f));
            }
            return;
        }
        if(lower.getVisibility()==VISIBLE) {
            lower.animate().cancel(); lower.setTranslationY(dp(32)); lower.setAlpha(0);
            lower.animate().translationY(0).alpha(1).setStartDelay(60).setDuration(220).setInterpolator(new OvershootInterpolator(1f));
        }
        sheet.setTranslationY(fromBelow ? dp(32) : -dp(48)); sheet.setAlpha(0);
        sheet.animate().translationY(0).alpha(1).setDuration(fromBelow ? 220 : 300)
                .setInterpolator(new OvershootInterpolator(fromBelow ? 1f : 1.4f));
    }
    @Override protected void onSizeChanged(int w,int h,int ow,int oh) { place(); }

    /** Sunken amber read-out in the style of the dot-matrix display. */
    static final class Readout extends View {
        private final Paint paint=new Paint();
        private final String text;
        Readout(Context context,String text) {
            super(context); this.text=PixelFont.normalize(text);
            setContentDescription(text); setFocusable(false);
        }
        @Override protected void onDraw(Canvas canvas) {
            float d=getResources().getDisplayMetrics().density;
            int u=Vga.unit(d), s=PixelFont.scale(16*d);
            while(s>1 && PixelFont.width(text,s)>getWidth()-8*u) s--;
            paint.setColor(Vga.DOT); canvas.drawRect(0,0,getWidth(),getHeight(),paint);
            paint.setColor(Vga.BLACK); canvas.drawRect(2*u,2*u,getWidth()-2*u,getHeight()-2*u,paint);
            paint.setColor(Vga.AMBER);
            PixelFont.draw(canvas,text,Math.round((getWidth()-PixelFont.width(text,s))/2f),
                    Math.round((getHeight()-PixelFont.CELL*s)/2f),s,paint);
        }
    }
}
