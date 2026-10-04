package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.view.View;
import android.widget.Button;
import android.widget.FrameLayout;
import android.widget.HorizontalScrollView;
import android.widget.LinearLayout;

// One default row. Both rows share the same safe-area container when expanded.
final class ControlMenu extends LinearLayout {
    final LinearLayout toolbar, auxiliary;
    final HorizontalScrollView panel;
    final Button keysButton;
    private final Controls controls;
    ControlMenu(Context context, Controls controls, Runnable importData) {
        super(context); this.controls=controls; setOrientation(VERTICAL);
        toolbar=new LinearLayout(context);
        addView(toolbar,new LinearLayout.LayoutParams(LayoutParams.MATCH_PARENT,LayoutParams.WRAP_CONTENT));
        String[] labels={"Enter","Esc","P","M","Y","N"};
        int[] codes={28,1,25,50,21,49};
        for (int i=0;i<labels.length;i++) key(toolbar,labels[i],codes[i]);
        button(toolbar,"Data").setOnClickListener(v->importData.run());
        keysButton=button(toolbar,"Keys");
        auxiliary=new LinearLayout(context);
        for (int i=0;i<8;i++) key(auxiliary,"F"+(i+1),59+i);
        for (int k=29;k<=54;k++) key(auxiliary,Character.toString((char)('A'+k-29)),Controls.make(k));
        panel=scroll(auxiliary); panel.setVisibility(GONE); addView(panel);
        keysButton.setOnClickListener(v->{
            boolean expanded=panel.getVisibility()==GONE;
            panel.setVisibility(expanded ? VISIBLE : GONE);
            keysButton.setSelected(expanded);
        });
    }
    private HorizontalScrollView scroll(LinearLayout row) {
        HorizontalScrollView view=new HorizontalScrollView(getContext());
        view.setFillViewport(true); view.setHorizontalScrollBarEnabled(false);
        view.addView(row); return view;
    }
    private Button button(LinearLayout row, String label) {
        Button button=new Button(getContext()); button.setText(label); button.setTextSize(11);
        button.setMinWidth(0); button.setMinimumWidth(0); button.setPadding(0,0,0,0);
        button.setAlpha(.65f); button.setFocusable(false);
        int size=Math.round(44*getResources().getDisplayMetrics().density);
        row.addView(button,new LinearLayout.LayoutParams(row==toolbar ? 0 : size,size,
                row==toolbar ? 1 : 0)); return button;
    }
    private void key(LinearLayout row,String label,int code) {
        button(row,label).setOnClickListener(v->controls.tap(code));
    }
    static androidx.core.graphics.Insets interactiveInsets(androidx.core.view.WindowInsetsCompat insets) {
        return insets.getInsets(androidx.core.view.WindowInsetsCompat.Type.systemBars()
                | androidx.core.view.WindowInsetsCompat.Type.displayCutout()
                | androidx.core.view.WindowInsetsCompat.Type.systemGestures());
    }
    void safeInsets(int left,int top,int right) {
        FrameLayout.LayoutParams layout=(FrameLayout.LayoutParams)getLayoutParams();
        layout.setMargins(left,top,right,0); setLayoutParams(layout);
    }
}
