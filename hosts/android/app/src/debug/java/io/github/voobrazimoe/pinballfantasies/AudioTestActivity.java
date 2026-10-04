package io.github.voobrazimoe.pinballfantasies;

import android.app.Activity;
import android.os.Bundle;
import android.util.Log;

// Debug-only foreground harness: same Android focus/device adapter, isolated PCM,
// no engine, no imported assets and no intent extras enabling production hooks.
public final class AudioTestActivity extends Activity {
    private AndroidAudio audio;
    private boolean resumed, focused;
    static native long nativeTestAudio(int operation);
    @Override public void onCreate(Bundle state) {
        super.onCreate(state);
        System.loadLibrary("pinball_android");
        audio=new AndroidAudio(this,(enabled,route)->nativeTestAudio(route?2:enabled?1:0),
                message->Log.i("PinballFantasies",message));
        setContentView(new android.widget.TextView(this));
    }
    private void sync() { if(audio!=null) audio.eligible(resumed && focused); }
    @Override public void onResume() { super.onResume(); resumed=true; sync(); }
    @Override public void onPause() { resumed=false; sync(); super.onPause(); }
    @Override public void onWindowFocusChanged(boolean value) { super.onWindowFocusChanged(value); focused=value; sync(); }
    @Override public void onDestroy() { audio.close(); super.onDestroy(); }
    void injectFocus(int event) { audio.focus(event); }
    void injectRoute() { audio.route(); }
    void injectNoisy() {
        // This is a protected system broadcast. Inject receiver delivery rather
        // than attempting to send it under the application UID.
        try {
            java.lang.reflect.Field field=AndroidAudio.class.getDeclaredField("noisy");
            field.setAccessible(true);
            ((android.content.BroadcastReceiver)field.get(audio)).onReceive(this,
                    new android.content.Intent(android.media.AudioManager.ACTION_AUDIO_BECOMING_NOISY));
        } catch (ReflectiveOperationException failure) { throw new AssertionError(failure); }
    }
    void injectBackground(boolean value) { resumed=!value; sync(); }
}
