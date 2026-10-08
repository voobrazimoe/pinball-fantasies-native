package io.github.voobrazimoe.pinballfantasies;

import android.app.Instrumentation;
import android.app.Activity;
import android.os.Bundle;
import android.view.MotionEvent;
import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.*;

/** Runs against the packaged Go library. All candidate bytes are invented and invalid. */
public final class ImportInstrumentation extends Instrumentation {
    private static native boolean nativeAudioSmoke();
    @Override public void onCreate(Bundle arguments) { super.onCreate(arguments); start(); }
    static void check(boolean value) { if (!value) throw new AssertionError(); }
    private void testFirstRunShell() {
        runOnMainSync(() -> {
            android.view.View menu = new android.view.View(getTargetContext());
            android.view.View touch = new android.view.View(getTargetContext());
            FirstRunShell shell = new FirstRunShell(getTargetContext(), () -> {});
            shell.present(false,false,false,menu,touch);
            check(shell.getVisibility()==android.view.View.VISIBLE && shell.button.isEnabled());
            check(menu.getVisibility()==android.view.View.GONE && touch.getVisibility()==android.view.View.GONE);
            shell.present(false,true,false,menu,touch);
            check(!shell.button.isEnabled() && shell.button.getText().toString().equals("Importing…"));
            shell.present(false,false,false,menu,touch);
            check(shell.button.isEnabled() && shell.button.getText().toString().equals("Import DOS folder"));
            shell.present(true,false,false,menu,touch);
            check(shell.getVisibility()==android.view.View.GONE && menu.getVisibility()==android.view.View.VISIBLE
                    && touch.getVisibility()==android.view.View.VISIBLE);
            shell.present(true,false,true,menu,touch);
            check(touch.getVisibility()==android.view.View.GONE);
            check(shell.demoButton==null);
            FirstRunShell choice = new FirstRunShell(getTargetContext(), () -> {}, () -> {});
            choice.present(false,false,false,menu,touch);
            check(choice.button.isEnabled() && choice.demoButton.isEnabled()
                    && choice.demoButton.getText().toString().equals("Play 10-minute demo"));
            choice.startingDemo = true;
            choice.present(false,true,false,menu,touch);
            check(!choice.button.isEnabled() && !choice.demoButton.isEnabled()
                    && choice.demoButton.getText().toString().equals("Starting demo…")
                    && choice.button.getText().toString().equals("Import DOS folder"));
            try { check(getTargetContext().getAssets().list("demo").length == 6); }
            catch (IOException failure) { throw new AssertionError(failure); }
        });
    }
    private void testOverlay() {
        runOnMainSync(() -> {
            List<String> events=new ArrayList<>();
            List<Runnable> releases=new ArrayList<>();
            Controls c=new Controls((kind,a,b)->events.add(kind+":"+a+":"+b),
                    (ms,release)->releases.add(release)); c.enabled=true; c.plungerAvailable=true;
            ControlOverlay view=new ControlOverlay(getTargetContext(),c); view.layout(0,0,1000,1000);
            view.viewport(new int[]{0,0,1000,1000,1000,1000,320,609});
            motion(view,MotionEvent.ACTION_DOWN,new int[]{17},new float[]{c.layout.left.cx()},new float[]{c.layout.left.cy()});
            motion(view,MotionEvent.ACTION_POINTER_DOWN | (1<<8),new int[]{17,91},new float[]{c.layout.left.cx(),c.layout.right.cx()},new float[]{c.layout.left.cy(),c.layout.right.cy()});
            check(c.held[0] && c.held[1]);
            motion(view,MotionEvent.ACTION_POINTER_UP | (1<<8),new int[]{17,91},new float[]{c.layout.left.cx(),c.layout.right.cx()},new float[]{c.layout.left.cy(),c.layout.right.cy()});
            check(c.held[0] && !c.held[1]);
            motion(view,MotionEvent.ACTION_CANCEL,new int[]{17},new float[]{c.layout.left.cx()},new float[]{c.layout.left.cy()});
            check(!c.held[0] && c.pointers.isEmpty());
            events.clear();
            motion(view,MotionEvent.ACTION_DOWN,new int[]{37},new float[]{900},new float[]{300});
            motion(view,MotionEvent.ACTION_MOVE,new int[]{37},new float[]{900},new float[]{425});
            motion(view,MotionEvent.ACTION_UP,new int[]{37},new float[]{900},new float[]{425});
            check(events.equals(List.of("5:16:0","4:0:0")));
            check(!c.pulling());
            events.clear();
            motion(view,MotionEvent.ACTION_DOWN,new int[]{38},new float[]{500},new float[]{300});
            check(events.isEmpty());
            motion(view,MotionEvent.ACTION_UP,new int[]{38},new float[]{500},new float[]{300});
            check(events.equals(List.of("0:3:1","1:57:0")));
            releases.remove(0).run(); check(!c.held[3]);
            testMenu(c);
        });
    }
    private void testMenu(Controls c) {
        android.content.Context context=getTargetContext();
        android.widget.FrameLayout root=new android.widget.FrameLayout(context);
        ControlOverlay overlay=new ControlOverlay(context,c);
        root.addView(overlay,new android.widget.FrameLayout.LayoutParams(-1,-1));
        ControlMenu menu=new ControlMenu(context,c);
        root.addView(menu,new android.widget.FrameLayout.LayoutParams(-1,-1));
        overlay.state=menu.state;
        for(int[] shape:new int[][]{{400,900,0,80,0,24},{360,640,12,48,24,16},{900,400,80,0,40,24}}) {
            float density=context.getResources().getDisplayMetrics().density;
            for(int i=0;i<shape.length;i++) shape[i]=Math.round(shape[i]*density);
            androidx.core.view.WindowInsetsCompat insets=new androidx.core.view.WindowInsetsCompat.Builder()
                    .setInsets(androidx.core.view.WindowInsetsCompat.Type.displayCutout(),
                            androidx.core.graphics.Insets.of(shape[2],shape[3],shape[4],0))
                    .setInsets(androidx.core.view.WindowInsetsCompat.Type.systemBars(),
                            androidx.core.graphics.Insets.of(0,16,0,shape[5]))
                    .setInsets(androidx.core.view.WindowInsetsCompat.Type.systemGestures(),
                            androidx.core.graphics.Insets.of(8,0,8,8)).build();
            androidx.core.graphics.Insets safe=ControlMenu.interactiveInsets(insets);
            check(safe.left==Math.max(8,shape[2]) && safe.top==Math.max(16,shape[3])
                    && safe.right==Math.max(8,shape[4]) && safe.bottom==shape[5]);
            menu.safeInsets(safe.left,safe.top,safe.right,safe.bottom);
            overlay.safeArea(safe.left,safe.top,safe.right,safe.bottom);
            menu.snapshot(SemanticUi.PLAYING,1,4); measure(root,shape[0],shape[1]);
            int[] viewport={shape[2],safe.top,shape[0]-safe.left-safe.right,shape[1]-safe.top-safe.bottom,
                    shape[0],shape[1],320,shape[0]<shape[1]?609:240};
            menu.viewport(viewport); overlay.viewport(viewport); measure(root,shape[0],shape[1]);
            check(menu.menu.getTop()>=safe.top+viewport[3]*33f/viewport[7]);
            check(menu.sheet.getVisibility()==android.view.View.GONE);
            check(c.layout.safe.top==safe.top && c.labelX(Controls.RIGHT)==c.layout.right.cx());
            check(!menu.menu.isFocusable());
            check(menu.menu.getTop()>=safe.top && menu.menu.getBottom()<=shape[1]-safe.bottom);
            check(menu.menu.getLeft()>=safe.left && menu.menu.getRight()<=shape[0]-safe.right);
            menu.menu.performClick(); measure(root,shape[0],shape[1]);
            check(menu.sheet.getVisibility()==android.view.View.VISIBLE);
            check(menu.sheet.getLeft()>=safe.left && menu.sheet.getRight()<=shape[0]-safe.right);
            check(menu.sheet.getTop()>=safe.top && menu.sheet.getBottom()<=shape[1]-safe.bottom);
            check(!c.panel.overlaps(c.layout.left) && !c.panel.overlaps(c.layout.right));
            check(!hasScroll(menu.sheet));
            check(findButton(menu.sheet,"Data / Import DOS folder")==null);
            findButton(menu.sheet,"Advanced keyboard").performClick(); measure(root,shape[0],shape[1]);
            check(findButton(menu.sheet,"F8")!=null && findButton(menu.sheet,"Z")!=null);
            check(menu.dismiss()); measure(root,shape[0],shape[1]);
            check(menu.sheet.getVisibility()==android.view.View.GONE && c.panel==null);
            eventsForHidden(root,c);
            menu.snapshot(SemanticUi.SELECTOR,0,0); measure(root,shape[0],shape[1]);
            check(findButton(menu.sheet,"Party Land")!=null && findButton(menu.sheet,"Options")!=null);
            check(!hasScroll(menu.sheet));
            check(menu.sheet.getTop()>safe.top+viewport[3]*.35f);
            android.widget.Button first=findButton(menu.sheet,"Party Land");
            for(String label:SemanticUi.TABLES) check(!findButton(menu.sheet,label).isFocusable());
            for(String label:SemanticUi.TABLES) {
                android.widget.Button cell=findButton(menu.sheet,label);
                check(cell.getWidth()==first.getWidth() && cell.getHeight()==first.getHeight());
            }
            check(menu.sheet.getTop()>=safe.top && menu.sheet.getBottom()<=shape[1]-safe.bottom);
            for(int mode:new int[]{SemanticUi.ATTRACT,SemanticUi.OPTIONS,SemanticUi.PAUSED,SemanticUi.QUESTION}) {
                menu.snapshot(mode,1,0); measure(root,shape[0],shape[1]);
                check(!hasScroll(menu.sheet));
                check(menu.sheet.getTop()>safe.top+viewport[3]*.35f);
                check(menu.sheet.getTop()>=safe.top && menu.sheet.getBottom()<=shape[1]-safe.bottom);
            }
            menu.snapshot(SemanticUi.INITIALS,1,0); measure(root,shape[0],shape[1]);
            check(findButton(menu.sheet,"A")!=null && findButton(menu.sheet,"1")==null);
            check(HardwareKeyboard.qualifies(true,false,android.view.InputDevice.KEYBOARD_TYPE_ALPHABETIC,
                    android.view.InputDevice.SOURCE_KEYBOARD));
            check(!HardwareKeyboard.qualifies(true,true,android.view.InputDevice.KEYBOARD_TYPE_ALPHABETIC,
                    android.view.InputDevice.SOURCE_KEYBOARD));
            check(!HardwareKeyboard.qualifies(true,false,android.view.InputDevice.KEYBOARD_TYPE_NONE,
                    android.view.InputDevice.SOURCE_GAMEPAD));
            check(!HardwareKeyboard.qualifies(false,false,android.view.InputDevice.KEYBOARD_TYPE_ALPHABETIC,
                    android.view.InputDevice.SOURCE_KEYBOARD));
            for(int mode:new int[]{SemanticUi.SELECTOR,SemanticUi.SELECTOR_TEXT,SemanticUi.ATTRACT,
                    SemanticUi.OPTIONS,SemanticUi.PAUSED,SemanticUi.QUESTION,SemanticUi.PLAYING,SemanticUi.INITIALS}) {
                menu.snapshot(mode,1,4); menu.keyboardMode(true); measure(root,shape[0],shape[1]);
                check(menu.getVisibility()==android.view.View.GONE && c.panel==null && c.menu==null);
                c.key(29,true,0); check(c.keys[29]); c.key(29,false,0);
                menu.keyboardMode(false); measure(root,shape[0],shape[1]);
                check(menu.getVisibility()==android.view.View.VISIBLE && menu.state.mode==mode);
                if(mode==SemanticUi.INITIALS) check(findButton(menu.sheet,"Q")!=null && findButton(menu.sheet,"1")==null);
                if(mode==SemanticUi.PLAYING) check(c.gameplay && !c.keyboardMode);
            }
            menu.menu.performClick(); measure(root,shape[0],shape[1]); check(!hasScroll(menu.sheet));
            check(menu.dismiss()); measure(root,shape[0],shape[1]);
            menu.snapshot(SemanticUi.ENTRY_WAIT,1,0); measure(root,shape[0],shape[1]);
            check(menu.sheet.getVisibility()==android.view.View.GONE && !c.gameplay);
        }
    }
    private static boolean hasScroll(android.view.View v) {
        if(v instanceof android.widget.ScrollView) return true;
        if(v instanceof android.view.ViewGroup) {
            android.view.ViewGroup g=(android.view.ViewGroup)v;
            for(int i=0;i<g.getChildCount();i++) if(hasScroll(g.getChildAt(i))) return true;
        }
        return false;
    }
    private static void eventsForHidden(android.view.View root,Controls c) {
        float x=c.layout.left.cx(),y=c.layout.left.cy();
        MotionEvent down=MotionEvent.obtain(0,1,MotionEvent.ACTION_DOWN,x,y,0);
        try { check(root.dispatchTouchEvent(down)); check(c.held[Controls.LEFT]); } finally { down.recycle(); }
        MotionEvent cancel=MotionEvent.obtain(0,2,MotionEvent.ACTION_CANCEL,x,y,0);
        try { root.dispatchTouchEvent(cancel); check(!c.held[Controls.LEFT]); } finally { cancel.recycle(); }
    }
    private static android.widget.Button findButton(android.view.View view,String text) {
        if(view instanceof android.widget.Button && ((android.widget.Button)view).getText().toString().equals(text))
            return (android.widget.Button)view;
        if(view instanceof android.view.ViewGroup) {
            android.view.ViewGroup group=(android.view.ViewGroup)view;
            for(int i=0;i<group.getChildCount();i++) {
                android.widget.Button found=findButton(group.getChildAt(i),text); if(found!=null) return found;
            }
        }
        return null;
    }

    private static void measure(android.view.View view,int w,int h) {
        for(int pass=0;pass<2;pass++) {
            view.measure(android.view.View.MeasureSpec.makeMeasureSpec(w,android.view.View.MeasureSpec.EXACTLY),
                    android.view.View.MeasureSpec.makeMeasureSpec(h,android.view.View.MeasureSpec.EXACTLY));
            view.layout(0,0,w,h);
        }
    }
    private static void motion(ControlOverlay view,int action,int[] ids,float[] xs,float[] ys) {
        MotionEvent.PointerProperties[] properties=new MotionEvent.PointerProperties[ids.length];
        MotionEvent.PointerCoords[] coords=new MotionEvent.PointerCoords[ids.length];
        for(int i=0;i<ids.length;i++) {
            properties[i]=new MotionEvent.PointerProperties(); properties[i].id=ids[i];
            properties[i].toolType=MotionEvent.TOOL_TYPE_FINGER;
            coords[i]=new MotionEvent.PointerCoords(); coords[i].x=xs[i]; coords[i].y=ys[i];
            coords[i].pressure=1; coords[i].size=1;
        }
        MotionEvent event=MotionEvent.obtain(0,1,action,ids.length,properties,coords,0,0,1,1,0,0,
                android.view.InputDevice.SOURCE_TOUCHSCREEN,0);
        try { view.dispatchTouchEvent(event); } finally { event.recycle(); }
    }
    private void testFocus() throws Exception {
        android.content.Intent intent=new android.content.Intent(getTargetContext(),AudioTestActivity.class);
        intent.addFlags(android.content.Intent.FLAG_ACTIVITY_NEW_TASK);
        AudioTestActivity activity=(AudioTestActivity)startActivitySync(intent);
        try {
            long end=android.os.SystemClock.uptimeMillis()+5000;
            while(AudioTestActivity.nativeTestAudio(4)==0 && android.os.SystemClock.uptimeMillis()<end) Thread.sleep(20);
            check(AudioTestActivity.nativeTestAudio(4)>0); // real top-Activity AudioManager grant
            consumeFresh();
            long before=AudioTestActivity.nativeTestAudio(5);
            runOnMainSync(()->activity.injectFocus(-2));
            check((AudioTestActivity.nativeTestAudio(5)&1)==0 && AudioTestActivity.nativeTestAudio(5)>before);
            Thread.sleep(150); long consumed=AudioTestActivity.nativeTestAudio(3);
            Thread.sleep(50); check(AudioTestActivity.nativeTestAudio(3)==consumed);
            runOnMainSync(()->activity.injectFocus(1)); consumeFresh();
            long opens=AudioTestActivity.nativeTestAudio(4);
            runOnMainSync(activity::injectRoute);
            end=android.os.SystemClock.uptimeMillis()+5000;
            while(AudioTestActivity.nativeTestAudio(4)<=opens && android.os.SystemClock.uptimeMillis()<end) Thread.sleep(20);
            check(AudioTestActivity.nativeTestAudio(4)>opens); consumeFresh();
            opens=AudioTestActivity.nativeTestAudio(4);
            runOnMainSync(activity::injectNoisy);
            end=android.os.SystemClock.uptimeMillis()+5000;
            while(AudioTestActivity.nativeTestAudio(4)<=opens && android.os.SystemClock.uptimeMillis()<end) Thread.sleep(20);
            check(AudioTestActivity.nativeTestAudio(4)>opens && (AudioTestActivity.nativeTestAudio(5)&1)!=0);
            consumeFresh();
            runOnMainSync(()->activity.injectBackground(true));
            check((AudioTestActivity.nativeTestAudio(5)&1)==0);
            runOnMainSync(()->activity.injectFocus(1)); check((AudioTestActivity.nativeTestAudio(5)&1)==0);
            runOnMainSync(()->activity.injectBackground(false)); consumeFresh();
            runOnMainSync(()->activity.injectFocus(-1));
            runOnMainSync(()->activity.injectFocus(1)); check((AudioTestActivity.nativeTestAudio(5)&1)==0);
        } finally { runOnMainSync(activity::finish); waitForIdleSync(); }
    }
    private static void consumeFresh() throws Exception {
        long baseline=AudioTestActivity.nativeTestAudio(3), end=android.os.SystemClock.uptimeMillis()+5000;
        while(android.os.SystemClock.uptimeMillis()<end) {
            if(AudioTestActivity.nativeTestAudio(3)>baseline) return;
            Thread.sleep(20);
        }
        throw new AssertionError("A5 fresh PCM consumption timed out");
    }
    @Override public void onStart() {
        Bundle result = new Bundle();
        File root = null;
        long session = 0;
        int outcome = Activity.RESULT_CANCELED;
        try {
            root = Files.createTempDirectory(getTargetContext().getNoBackupFilesDir().toPath(), "a2-test-").toFile();
            File data = new File(root, "Data"), state = new File(root, "State.validation-test");
            Files.createDirectories(data.toPath()); Files.createDirectories(state.toPath());
            PinballActivity.nativeDiagnostics(true);
            session = PinballActivity.nativeOpen();
            String missing = PinballActivity.nativeEngine(session, 0,
                    data.getAbsolutePath().getBytes(StandardCharsets.UTF_8),
                    state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
            check(missing != null && !missing.isEmpty() && missing.length() < 1024);
            for (String name : DataImport.REQUIRED) Files.write(new File(data, name).toPath(), new byte[]{1,2,3});
            String malformed = PinballActivity.nativeEngine(session, 0,
                    data.getAbsolutePath().getBytes(StandardCharsets.UTF_8),
                    state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
            check(malformed != null && !malformed.isEmpty() && malformed.length() < 1024);
            String uri = PinballActivity.nativeEngine(session, 0, "content://test".getBytes(StandardCharsets.UTF_8),
                    state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
            check(uri != null && uri.contains("filesystem"));
            String boot = PinballActivity.nativeEngine(session, 1,
                    data.getAbsolutePath().getBytes(StandardCharsets.UTF_8),
                    state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
            check(boot != null && !boot.isEmpty());
            PinballActivity.nativeClose(session);
            String closed = PinballActivity.nativeEngine(session, 0,
                    data.getAbsolutePath().getBytes(StandardCharsets.UTF_8),
                    state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
            check("Activity closed".equals(closed));
            // No-data JNI calls must remain harmless across focus/pause/stale sessions.
            PinballActivity.nativeActive(session,true,true);
            for (int kind=0;kind<5;kind++) PinballActivity.nativeInput(session,kind,0,1);
            testFirstRunShell();
            testOverlay();
            check(nativeAudioSmoke());
            testFocus();
            result.putString("stream", "PASS: A5 foreground focus, injected interruption/gain, background/resume and deferred route reopen\nPASS: A4 packaged Oboe 48 kHz stereo opens, callbacks, synthetic PCM, pause/resume and close\nPASS: A3 real Android MotionEvent pointer/cancel/plunger dispatch, semantic sheets and synthetic safe cutout layout and no-data lifecycle/input JNI\nPASS: packaged ABI 1 JNI missing/malformed rejection, bounded error, URI/session guards\n");
            outcome = Activity.RESULT_OK;
        } catch (Throwable failure) {
            result.putString("stream", "FAIL: " + failure + "\n");
        } finally {
            PinballActivity.nativeClose(session);
            if (root != null) try { DataImport.remove(root); } catch (IOException ignored) { }
        }
        finish(outcome, result);
    }
}
