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
    @Override public void onCreate(Bundle arguments) { super.onCreate(arguments); start(); }
    static void check(boolean value) { if (!value) throw new AssertionError(); }
    private void testOverlay() {
        runOnMainSync(() -> {
            List<String> events=new ArrayList<>();
            Controls c=new Controls((kind,a,b)->events.add(kind+":"+a+":"+b)); c.enabled=true;
            ControlOverlay view=new ControlOverlay(getTargetContext(),c); view.layout(0,0,1000,1000);
            motion(view,MotionEvent.ACTION_DOWN,new int[]{17},new float[]{100},new float[]{800});
            motion(view,MotionEvent.ACTION_POINTER_DOWN | (1<<8),new int[]{17,91},new float[]{100,600},new float[]{800,800});
            check(c.held[0] && c.held[1]);
            motion(view,MotionEvent.ACTION_POINTER_UP | (1<<8),new int[]{17,91},new float[]{100,600},new float[]{800,800});
            check(c.held[0] && !c.held[1]);
            motion(view,MotionEvent.ACTION_CANCEL,new int[]{17},new float[]{100},new float[]{800});
            check(!c.held[0] && c.pointers.isEmpty());
            events.clear();
            motion(view,MotionEvent.ACTION_DOWN,new int[]{37},new float[]{900},new float[]{700});
            motion(view,MotionEvent.ACTION_MOVE,new int[]{37},new float[]{900},new float[]{825});
            motion(view,MotionEvent.ACTION_UP,new int[]{37},new float[]{900},new float[]{825});
            check(events.equals(List.of("3:64:0","4:0:0")));
        });
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
    @Override public void onStart() {
        Bundle result = new Bundle();
        File root = null;
        long session = 0;
        int outcome = Activity.RESULT_CANCELED;
        try {
            root = Files.createTempDirectory(getTargetContext().getNoBackupFilesDir().toPath(), "a2-test-").toFile();
            File data = new File(root, "Data"), state = new File(root, "State.validation-test");
            Files.createDirectories(data.toPath()); Files.createDirectories(state.toPath());
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
            result.putString("stream", "PASS: A3 real Android MotionEvent pointer/cancel/plunger dispatch and no-data lifecycle/input JNI\nPASS: packaged ABI 1 JNI missing/malformed rejection, bounded error, URI/session guards\n");
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
