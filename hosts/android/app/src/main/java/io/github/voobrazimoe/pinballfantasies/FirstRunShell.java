package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.graphics.Color;
import android.view.Gravity;
import android.view.View;
import android.widget.*;

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
        content.setGravity(Gravity.CENTER);
        int pad = Math.round(24 * getResources().getDisplayMetrics().density);
        content.setPadding(pad,pad,pad,pad);
        addContent(content,"Pinball Fantasies",26);
        addContent(content,demo == null ? "Select the folder containing your original DOS game files."
                : "Import the full game from your original DOS folder, or play the official 10-minute Party Land demo.",16);
        button = new Button(context);
        button.setText("Import DOS folder");
        button.setOnClickListener(v -> request.run());
        content.addView(button);
        if (demo != null) {
            demoButton = new Button(context);
            demoButton.setText("Play 10-minute demo");
            demoButton.setOnClickListener(v -> demo.run());
            content.addView(demoButton);
        } else demoButton = null;
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
        button.setText(busy && !startingDemo ? "Importing…" : "Import DOS folder");
        if (demoButton != null) {
            demoButton.setEnabled(!busy && !loaded);
            demoButton.setText(busy && startingDemo ? "Starting demo…" : "Play 10-minute demo");
        }
    }
}
