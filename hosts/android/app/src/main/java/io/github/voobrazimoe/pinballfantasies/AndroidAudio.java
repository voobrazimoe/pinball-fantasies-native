package io.github.voobrazimoe.pinballfantasies;

import android.content.*;
import android.media.*;
import android.os.*;
import java.util.function.Consumer;

/** All methods and Android notifications run on the UI looper. No stream operations. */
final class AndroidAudio {
    interface Output { void publish(boolean enabled, boolean route); }
    private final Context context;
    private final AudioManager manager;
    private final AudioFocusRequest request;
    private final AudioPolicy policy;
    private final Output output;
    private final Consumer<String> log;
    private final Handler handler = new Handler(Looper.getMainLooper());
    private boolean registered, closed;
    private long losses, recoveries;
    private final AudioDeviceCallback devices = new AudioDeviceCallback() {
        @Override public void onAudioDevicesAdded(AudioDeviceInfo[] list) { deviceChange(list); }
        @Override public void onAudioDevicesRemoved(AudioDeviceInfo[] list) { deviceChange(list); }
    };
    private final BroadcastReceiver noisy = new BroadcastReceiver() {
        @Override public void onReceive(Context c, Intent intent) {
            if (AudioManager.ACTION_AUDIO_BECOMING_NOISY.equals(intent.getAction())) {
                log.accept("A5_BECOMING_NOISY"); route();
            }
        }
    };
    AndroidAudio(Context context, Output output, Consumer<String> log) {
        this.context=context; this.output=output; this.log=log;
        manager=(AudioManager)context.getSystemService(Context.AUDIO_SERVICE);
        request=new AudioFocusRequest.Builder(AudioManager.AUDIOFOCUS_GAIN)
                .setAudioAttributes(new AudioAttributes.Builder().setUsage(AudioAttributes.USAGE_GAME)
                        .setContentType(AudioAttributes.CONTENT_TYPE_MUSIC).build())
                .setWillPauseWhenDucked(true).setAcceptsDelayedFocusGain(false)
                .setOnAudioFocusChangeListener(this::focus,handler).build();
        policy=new AudioPolicy(new AudioPolicy.Host() {
            public boolean request() {
                boolean granted=manager.requestAudioFocus(request)==AudioManager.AUDIOFOCUS_REQUEST_GRANTED;
                log.accept("A5_FOCUS_REQUEST granted="+granted);
                return granted;
            }
            public void abandon() {
                manager.abandonAudioFocusRequest(request);
                // Dedicated handler: discard notifications already queued for the retired grant.
                handler.removeCallbacksAndMessages(null);
            }
            public void output(boolean enabled) { publish(enabled); }
        });
    }
    void eligible(boolean value) { policy.eligible(value); }
    void focus(int event) {
        if (closed) return;
        if (event<0) ++losses; else if(event==AudioManager.AUDIOFOCUS_GAIN) ++recoveries;
        log.accept("A5_FOCUS event="+event+" losses="+losses+" recoveries="+recoveries);
        policy.focus(event);
    }
    void route() {
        if (closed || !registered) return;
        log.accept("A5_ROUTE"); output.publish(false,true);
    }
    private void deviceChange(AudioDeviceInfo[] list) {
        if (closed || !registered) return;
        boolean sink=false;
        for (AudioDeviceInfo device:list) if(device.isSink()) {
            sink=true; log.accept("A5_DEVICE id="+device.getId()+" type="+device.getType());
        }
        if (sink) route();
    }
    private void publish(boolean enabled) {
        output.publish(enabled,false);
        if (enabled==registered) return;
        registered=enabled;
        if (enabled) {
            manager.registerAudioDeviceCallback(devices,handler);
            IntentFilter filter=new IntentFilter(AudioManager.ACTION_AUDIO_BECOMING_NOISY);
            if(Build.VERSION.SDK_INT>=33) context.registerReceiver(noisy,filter,Context.RECEIVER_NOT_EXPORTED);
            else context.registerReceiver(noisy,filter);
        } else {
            manager.unregisterAudioDeviceCallback(devices); context.unregisterReceiver(noisy);
        }
    }
    void close() { if(closed) return; policy.close(); closed=true; }
}
