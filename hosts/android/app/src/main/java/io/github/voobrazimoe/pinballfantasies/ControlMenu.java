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
    final VgaPanel sheet;
    private LinearLayout content;
    private FrameLayout selector;
    private boolean scrollable;
    private boolean opened, advanced;
    private boolean pulled;
    private float pullStart;
    private String shown="";
    private boolean enter;
    private int safeLeft,safeTop,safeRight,safeBottom;
    ControlMenu(Context context, Controls controls) {
        super(context); this.controls=controls;
        sheet=new VgaPanel(context);
        sheet.setOnTouchListener((v,e)->false);
        addView(sheet,new FrameLayout.LayoutParams(LayoutParams.MATCH_PARENT,LayoutParams.WRAP_CONTENT,Gravity.BOTTOM));
        VgaButton tab=new VgaButton(context,"Menu",VgaButton.TAB); tab.glyph="▼";
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
    private Button button(LinearLayout row,String label,Runnable action) {
        VgaButton b=new VgaButton(getContext(),label,label.equals("Play")?VgaButton.PRIMARY:VgaButton.PANEL);
        b.hint=hint(label); b.pulse=label.equals("Play");
        LinearLayout.LayoutParams lp=new LinearLayout.LayoutParams(0,dp(48),1);
        if(row.getChildCount()>0) lp.leftMargin=dp(8);
        row.addView(b,lp);
        b.setOnClickListener(v->action.run()); return b;
    }
    /** The DOS key each touch action stands for, shown as a chip on its button. */
    static String hint(String label) {
        switch(label) {
            case "Pause": return "P";
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
        LinearLayout.LayoutParams lp=new LinearLayout.LayoutParams(-1,dp(48));
        if(parent.getChildCount()>0) lp.topMargin=dp(8);
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
            if(state.mode==SemanticUi.PLAYING || state.mode==SemanticUi.ATTRACT) sheet.title(tableName(),"Menu");
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
    private void rebuild() {
        sheet.removeAllViews();
        selector=null; scrollable=advanced || (!opened && state.mode==SemanticUi.INITIALS);
        content=new LinearLayout(getContext()); content.setOrientation(LinearLayout.VERTICAL);
        content.setPadding(dp(8),dp(8),dp(8),dp(8));
        if(scrollable) { ScrollView scroll=new ScrollView(getContext()); scroll.addView(content); sheet.addView(scroll); }
        else sheet.addView(content);
        if (opened || advanced) {
            if (advanced) {
                button(row(content),"Close",()->{opened=advanced=false; rebuild();});
                String[] labels={"Enter","Esc","P","M","Y","N","↑","↓","←","→","Space"};
                int[] codes={28,1,25,50,21,49,72,80,75,77,57};
                for(int i=0;i<labels.length;i+=4) {
                    LinearLayout r=row(content);
                    for(int j=i;j<Math.min(i+4,labels.length);j++) { final int code=codes[j]; button(r,labels[j],()->controls.tap(code)); }
                }
                for(int i=0;i<8;i+=4) { LinearLayout r=row(content); for(int j=i;j<i+4;j++) {
                    final int code=59+j; button(r,"F"+(j+1),()->controls.tap(code));
                }}
                letters(content);
            } else {
                String[] utilities=state.utilities();
                LinearLayout r=row(content);
                button(r,"Close",()->{opened=advanced=false; rebuild();});
                for(int i=0;i<utilities.length;i++) {
                    if((i+1)%2==0) r=row(content);
                    String label=utilities[i];
                    button(r,label,()->{
                        if(label.equals("Advanced keyboard")) {advanced=true; opened=false; rebuild();}
                        else action(label);
                    });
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
            r.addView(count,new LinearLayout.LayoutParams(0,dp(48),2));
            ((LinearLayout.LayoutParams)count.getLayoutParams()).leftMargin=dp(8);
            button(r,"+",()->{state.players(1); rebuild();}).setContentDescription("More players");
            button(row(content),"Play",()->action("Play"));
        } else if(state.mode==SemanticUi.INITIALS) letters(content);
        else if(state.mode!=SemanticUi.STARTUP) {
            String[] actions=state.actions();
            for(int i=0;i<actions.length;i+=2) {
                LinearLayout r=row(content);
                for(int j=i;j<Math.min(i+2,actions.length);j++) {String label=actions[j]; button(r,label,()->action(label));}
            }
        }
        title();
        sheet.setVisibility(selector==null && content.getChildCount()==0 ? GONE : VISIBLE);
        // The selector's text pages keep the same cards, so they do not replay the entrance.
        int mode=state.mode==SemanticUi.SELECTOR_TEXT ? SemanticUi.SELECTOR : state.mode;
        String key=mode+(opened?"o":"")+(advanced?"a":"");
        enter=sheet.getVisibility()==VISIBLE && !key.equals(shown);
        shown=sheet.getVisibility()==VISIBLE ? key : "";
        place();
    }
    private void letters(LinearLayout content) {
        for(String line:new String[]{"QWERTYUIOP","ASDFGHJKL","ZXCVBNM"}) {
            LinearLayout r=row(content);
            for(char letter:line.toCharArray()) {String label=String.valueOf(letter); button(r,label,()->action(label));}
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
        panel.gravity=Gravity.TOP|Gravity.LEFT; panel.width=Math.round(grid.bounds.width()); panel.height=height;
        int x=Math.round(grid.bounds.left);
        int end=getHeight()-safeBottom;
        if(controls.gameplay && !scrollable && controls.layout!=null)
            end=Math.round(Math.min(controls.layout.leftGuard.top,controls.layout.rightGuard.top));
        InteractionGeometry.Rect safe=new InteractionGeometry.Rect(safeLeft,safeTop,getWidth()-safeRight,end);
        InteractionGeometry.Rect frame=InteractionGeometry.viewport(viewport,getWidth(),getHeight());
        boolean contextual=!opened && !advanced && (state.mode==SemanticUi.SELECTOR || state.mode==SemanticUi.SELECTOR_TEXT
                || state.mode==SemanticUi.ATTRACT || state.mode==SemanticUi.OPTIONS || state.mode==SemanticUi.PAUSED
                || state.mode==SemanticUi.QUESTION);
        int y=contextual ? Math.round(InteractionGeometry.lowerTop(safe,frame,height,d)) : scrollable ? safeTop+Math.max(0,(end-safeTop-height)/2) : grid.topWithin(safeTop,end);
        panel.setMargins(x,y,0,0); sheet.setLayoutParams(panel);
        controls.panel=!controls.keyboardMode && sheet.getVisibility()==VISIBLE ? new InteractionGeometry.Rect(x,y,x+panel.width,y+height) : null;
        FrameLayout.LayoutParams icon=(FrameLayout.LayoutParams)menu.getLayoutParams();
        InteractionGeometry.Rect iconBounds=InteractionGeometry.mobileMenu(
                new InteractionGeometry.Rect(safeLeft,safeTop,getWidth()-safeRight,getHeight()-safeBottom),frame,viewport[7],d);
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
