package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.view.Gravity;
import android.view.View;
import android.widget.Button;
import android.widget.FrameLayout;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;

// In-Activity sheets never create a Window or change audio/lifecycle eligibility.
final class ControlMenu extends FrameLayout {
    final SemanticUi state=new SemanticUi();
    private final Controls controls;
    final Button menu;
    final FrameLayout sheet;
    private LinearLayout content;
    private FrameLayout selector;
    private boolean scrollable;
    private boolean opened, advanced;
    private int safeLeft,safeTop,safeRight,safeBottom;
    ControlMenu(Context context, Controls controls) {
        super(context); this.controls=controls;
        sheet=new FrameLayout(context);
        sheet.setBackgroundColor(0xdd111111); sheet.setOnTouchListener((v,e)->false);
        addView(sheet,new FrameLayout.LayoutParams(LayoutParams.MATCH_PARENT,LayoutParams.WRAP_CONTENT,Gravity.BOTTOM));
        menu=new Button(context); menu.setText("⋮"); menu.setContentDescription("Mobile menu");
        menu.setMinWidth(0); menu.setMinimumWidth(0); menu.setPadding(0,0,0,0); menu.setAlpha(.65f);
        addView(menu,new FrameLayout.LayoutParams(dp(44),dp(44),Gravity.TOP|Gravity.RIGHT));
        menu.setOnClickListener(v->{controls.cancel(); opened=!opened; advanced=false; rebuild();});
        setClipChildren(true); rebuild();
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
        Button b=new Button(getContext()); b.setText(label); b.setTextSize(14); b.setFocusable(false);
        normalize(b);
        LinearLayout.LayoutParams lp=new LinearLayout.LayoutParams(0,dp(48),1);
        if(row.getChildCount()>0) lp.leftMargin=dp(8);
        row.addView(b,lp);
        b.setOnClickListener(v->action.run()); return b;
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
    private void rebuild() {
        sheet.removeAllViews();
        selector=null; scrollable=advanced || state.mode==SemanticUi.INITIALS;
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
            for(String label:SemanticUi.TABLES) {
                Button b=new Button(getContext()); b.setText(label); b.setTextSize(14); normalize(b);
                b.setOnClickListener(v->action(label)); selector.addView(b);
            }
        } else if(state.mode==SemanticUi.ATTRACT) {
            LinearLayout r=row(content);
            button(r,"−",()->{state.players(-1); rebuild();});
            TextView count=new TextView(getContext()); count.setText("Players   "+state.players); count.setTextColor(0xffffffff);
            count.setGravity(Gravity.CENTER); r.addView(count,new LinearLayout.LayoutParams(0,dp(48),2));
            button(r,"+",()->{state.players(1); rebuild();});
            button(row(content),"PLAY",()->action("Play"));
        } else if(state.mode==SemanticUi.INITIALS) letters(content);
        else if(state.mode!=SemanticUi.STARTUP) {
            String[] actions=state.actions();
            for(int i=0;i<actions.length;i+=2) {
                LinearLayout r=row(content);
                for(int j=i;j<Math.min(i+2,actions.length);j++) {String label=actions[j]; button(r,label,()->action(label));}
            }
        }
        sheet.setVisibility(selector==null && content.getChildCount()==0 ? GONE : VISIBLE);
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
    private void normalize(Button b) {
        b.setMinWidth(0); b.setMinimumWidth(0); b.setMinHeight(0); b.setMinimumHeight(0);
        b.setPadding(dp(4),0,dp(4),0); b.setMaxLines(2); b.setSingleLine(false);
        b.setGravity(Gravity.CENTER); b.setAllCaps(false); b.setTextColor(0xff111111);
        // Remove theme's optical insets; each visible card fills its exact cell.
        android.graphics.drawable.GradientDrawable background=new android.graphics.drawable.GradientDrawable();
        background.setColor(0xffdddddd); background.setCornerRadius(dp(4)); b.setBackground(background);
    }
    private void place() {
        if(getWidth()==0 || getHeight()==0) return;
        float d=getResources().getDisplayMetrics().density;
        FrameLayout.LayoutParams panel=(FrameLayout.LayoutParams)sheet.getLayoutParams();
        InteractionGeometry.Panel grid=new InteractionGeometry.Panel(getWidth(),getHeight(),safeLeft,safeTop,
                safeRight,safeBottom,d,selector!=null,selector!=null?5:Math.max(1,content.getChildCount()*2));
        int height;
        if(selector!=null) {
            height=Math.round(grid.bounds.height());
            for(int i=0;i<5;i++) {
                InteractionGeometry.Rect cell=grid.cells[i];
                FrameLayout.LayoutParams lp=new FrameLayout.LayoutParams(Math.round(cell.width()),Math.round(cell.height()));
                lp.leftMargin=Math.round(cell.left-grid.bounds.left); lp.topMargin=Math.round(cell.top-grid.bounds.top);
                selector.getChildAt(i).setLayoutParams(lp);
            }
        } else {
            content.measure(MeasureSpec.makeMeasureSpec(Math.round(grid.bounds.width()),MeasureSpec.EXACTLY),
                    MeasureSpec.makeMeasureSpec(0,MeasureSpec.UNSPECIFIED));
            height=content.getMeasuredHeight();
            if(scrollable) height=Math.min(height,Math.max(0,getHeight()-safeTop-safeBottom-dp(64)));
        }
        panel.gravity=Gravity.TOP|Gravity.LEFT; panel.width=Math.round(grid.bounds.width()); panel.height=height;
        int x=Math.round(grid.bounds.left);
        int end=getHeight()-safeBottom;
        if(controls.gameplay && !scrollable && controls.layout!=null)
            end=Math.round(Math.min(controls.layout.leftGuard.top,controls.layout.rightGuard.top));
        int y=scrollable ? safeTop+Math.max(0,(end-safeTop-height)/2) : grid.topWithin(safeTop,end);
        panel.setMargins(x,y,0,0); sheet.setLayoutParams(panel);
        controls.panel=sheet.getVisibility()==VISIBLE ? new InteractionGeometry.Rect(x,y,x+panel.width,y+height) : null;
        FrameLayout.LayoutParams icon=(FrameLayout.LayoutParams)menu.getLayoutParams();
        icon.gravity=Gravity.TOP|Gravity.RIGHT; icon.setMargins(0,safeTop,safeRight,0); menu.setLayoutParams(icon);
        controls.menu=new InteractionGeometry.Rect(getWidth()-safeRight-dp(44),safeTop,getWidth()-safeRight,safeTop+dp(44));
    }
    @Override protected void onSizeChanged(int w,int h,int ow,int oh) { place(); }
}
