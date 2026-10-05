package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.graphics.Color;
import android.view.Gravity;
import android.view.View;
import android.widget.*;

/** Opaque no-data UI above the native surface. No engine controls live here. */
final class FirstRunShell extends FrameLayout {
    final Button button;
    FirstRunShell(Context context, Runnable request) {
        super(context);
        setBackgroundColor(Color.BLACK);
        setClickable(true);
        LinearLayout content = new LinearLayout(context);
        content.setOrientation(LinearLayout.VERTICAL);
        content.setGravity(Gravity.CENTER);
        int pad = Math.round(24 * getResources().getDisplayMetrics().density);
        content.setPadding(pad,pad,pad,pad);
        addContent(content,"Pinball Fantasies",26);
        addContent(content,"Select the folder containing your original DOS game files.",16);
        button = new Button(context);
        button.setText("Import DOS folder");
        button.setOnClickListener(v -> request.run());
        content.addView(button);
        addContent(content,"The files are validated and copied to private app storage.",13);
        addView(content,new FrameLayout.LayoutParams(-1,-2,Gravity.CENTER));
    }
    private void addContent(LinearLayout parent, String text, int size) {
        TextView label = new TextView(getContext());
        label.setText(text); label.setTextColor(Color.LTGRAY); label.setTextSize(size);
        label.setGravity(Gravity.CENTER); label.setPadding(0,12,0,12);
        parent.addView(label,new LinearLayout.LayoutParams(-1,-2));
    }
    void present(boolean loaded, boolean busy, boolean keyboard, View menu, View touch) {
        setVisibility(loaded ? GONE : VISIBLE);
        menu.setVisibility(loaded && !keyboard ? VISIBLE : GONE);
        touch.setVisibility(loaded && !keyboard ? VISIBLE : GONE);
        button.setEnabled(!busy && !loaded);
        button.setText(busy ? "Importing…" : "Import DOS folder");
    }
}
