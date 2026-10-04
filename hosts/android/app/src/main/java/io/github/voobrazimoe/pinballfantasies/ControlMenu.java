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
    private final Runnable importData;
    final Button menu;
    final ScrollView sheet;
    private boolean opened, advanced;
    private int safeLeft,safeTop,safeRight,safeBottom;
    ControlMenu(Context context, Controls controls, Runnable importData) {
        super(context); this.controls=controls; this.importData=importData;
        sheet=new ScrollView(context); sheet.setFillViewport(false);
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
        b.setMinWidth(0); b.setMinimumWidth(0); b.setPadding(dp(4),0,dp(4),0);
        row.addView(b,new LinearLayout.LayoutParams(0,dp(48),1));
        b.setOnClickListener(v->action.run()); return b;
    }
    private LinearLayout row(LinearLayout parent) {
        LinearLayout row=new LinearLayout(getContext()); parent.addView(row); return row;
    }
    private void action(String label) {
        controls.tap(state.code(label));
        if (opened) { opened=false; rebuild(); }
    }
    private void rebuild() {
        sheet.removeAllViews();
        LinearLayout content=new LinearLayout(getContext()); content.setOrientation(LinearLayout.VERTICAL);
        content.setPadding(dp(8),dp(4),dp(8),dp(4)); sheet.addView(content);
        if (opened || advanced) {
            button(row(content),"Close",()->{opened=advanced=false; rebuild();});
            if (advanced) {
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
            } else for(String label:state.utilities()) button(row(content),label,()->{
                if(label.equals("Advanced keyboard")) {advanced=true; opened=false; rebuild();}
                else if(label.startsWith("Data")) {dismiss(); importData.run();}
                else action(label);
            });
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
        sheet.setVisibility(content.getChildCount()==0 ? GONE : VISIBLE);
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
        FrameLayout.LayoutParams panel=(FrameLayout.LayoutParams)sheet.getLayoutParams();
        panel.setMargins(safeLeft,0,safeRight,safeBottom);
        // Sheets are bounded to the bottom half, leaving the top matrix visible.
        panel.height=Math.max(dp(48),Math.min(sheetContentHeight(),(getHeight()-safeTop-safeBottom)/2));
        sheet.setLayoutParams(panel);
        FrameLayout.LayoutParams icon=(FrameLayout.LayoutParams)menu.getLayoutParams();
        float frameHeight=getWidth()<getHeight() ? getWidth()*609f/320f : getHeight();
        int letterbox=Math.max(0,Math.round((getHeight()-frameHeight)/2));
        boolean topSpace=letterbox-safeTop>=dp(44);
        icon.gravity=Gravity.RIGHT|(topSpace ? Gravity.TOP : Gravity.BOTTOM);
        icon.setMargins(safeLeft,safeTop,safeRight,safeBottom+(topSpace?0:(sheet.getVisibility()==VISIBLE?panel.height:0)));
        menu.setLayoutParams(icon);
    }
    private int sheetContentHeight() {
        if(sheet.getChildCount()==0) return dp(48);
        sheet.getChildAt(0).measure(MeasureSpec.makeMeasureSpec(Math.max(0,getWidth()-safeLeft-safeRight),MeasureSpec.EXACTLY),
                MeasureSpec.makeMeasureSpec(0,MeasureSpec.UNSPECIFIED));
        return sheet.getChildAt(0).getMeasuredHeight();
    }
    @Override protected void onSizeChanged(int w,int h,int ow,int oh) { place(); }
}
