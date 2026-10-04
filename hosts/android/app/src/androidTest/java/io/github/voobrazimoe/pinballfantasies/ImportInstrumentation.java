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
    private void testOverlay() {
        runOnMainSync(() -> {
            List<String> events=new ArrayList<>();
            List<Runnable> releases=new ArrayList<>();
            Controls c=new Controls((kind,a,b)->events.add(kind+":"+a+":"+b),
                    (ms,release)->releases.add(release)); c.enabled=true;
            ControlOverlay view=new ControlOverlay(getTargetContext(),c); view.layout(0,0,1000,1000);
            motion(view,MotionEvent.ACTION_DOWN,new int[]{17},new float[]{100},new float[]{800});
            motion(view,MotionEvent.ACTION_POINTER_DOWN | (1<<8),new int[]{17,91},new float[]{100,600},new float[]{800,800});
            check(c.held[0] && c.held[1]);
            motion(view,MotionEvent.ACTION_POINTER_UP | (1<<8),new int[]{17,91},new float[]{100,600},new float[]{800,800});
            check(c.held[0] && !c.held[1]);
            motion(view,MotionEvent.ACTION_CANCEL,new int[]{17},new float[]{100},new float[]{800});
            check(!c.held[0] && c.pointers.isEmpty());
            events.clear();
            motion(view,MotionEvent.ACTION_DOWN,new int[]{37},new float[]{900},new float[]{300});
            motion(view,MotionEvent.ACTION_MOVE,new int[]{37},new float[]{900},new float[]{425});
            motion(view,MotionEvent.ACTION_UP,new int[]{37},new float[]{900},new float[]{425});
            check(events.equals(List.of("3:64:0","4:0:0")));
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
        ControlMenu menu=new ControlMenu(context,c,()->{});
        root.addView(menu,new android.widget.FrameLayout.LayoutParams(-1,-2,android.view.Gravity.TOP));
        check(menu.toolbar.getChildCount()==8 && menu.panel.getVisibility()==android.view.View.GONE);
        String[] labels={"Enter","Esc","P","M","Y","N","Data","Keys"};
        for(int i=0;i<8;i++) check(((android.widget.Button)menu.toolbar.getChildAt(i)).getText().toString().equals(labels[i]));
        check(menu.auxiliary.getChildCount()==34);
        for(int i=0;i<8;i++) check(((android.widget.Button)menu.auxiliary.getChildAt(i)).getText().toString().equals("F"+(i+1)));
        for(int i=0;i<26;i++) check(((android.widget.Button)menu.auxiliary.getChildAt(i+8)).getText().toString().equals(Character.toString((char)('A'+i))));
        for(int[] shape:new int[][]{{400,900,0,80,0,24},{360,640,12,48,24,16},{900,400,80,0,40,24}}) {
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
            menu.safeInsets(safe.left,safe.top,safe.right);
            measure(root,shape[0],shape[1]);
            int collapsed=menu.getHeight();
            check(menu.getTop()==safe.top && menu.getLeft()==safe.left && menu.getRight()==shape[0]-safe.right);
            menu.keysButton.performClick(); measure(root,shape[0],shape[1]);
            check(menu.panel.getVisibility()==android.view.View.VISIBLE && menu.getHeight()==collapsed*2);
            overlay.safeArea(safe.left,safe.top,safe.right,safe.bottom,menu.getBottom());
            check(c.hit(shape[0]/2,menu.getBottom()-1)==Controls.NONE);
            menu.keysButton.performClick(); measure(root,shape[0],shape[1]);
            check(menu.panel.getVisibility()==android.view.View.GONE && menu.getHeight()==collapsed);
            overlay.safeArea(safe.left,safe.top,safe.right,safe.bottom,menu.getBottom());
            check(c.hit(shape[0]/2,menu.getBottom()+1)==Controls.PENDING);
            check(c.hit(safe.left-1,shape[1]/2)==Controls.NONE);
        }
    }
    private static void measure(android.view.View view,int w,int h) {
        view.measure(android.view.View.MeasureSpec.makeMeasureSpec(w,android.view.View.MeasureSpec.EXACTLY),
                android.view.View.MeasureSpec.makeMeasureSpec(h,android.view.View.MeasureSpec.EXACTLY));
        view.layout(0,0,w,h);
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
            testOverlay();
            check(nativeAudioSmoke());
            testFocus();
            result.putString("stream", "PASS: A5 foreground focus, injected interruption/gain, background/resume and deferred route reopen\nPASS: A4 packaged Oboe 48 kHz stereo opens, callbacks, synthetic PCM, pause/resume and close\nPASS: A3 real Android MotionEvent pointer/cancel/plunger dispatch, Keys panel and synthetic safe cutout layout and no-data lifecycle/input JNI\nPASS: packaged ABI 1 JNI missing/malformed rejection, bounded error, URI/session guards\n");
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
