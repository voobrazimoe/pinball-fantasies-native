package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.hardware.input.InputManager;
import android.view.InputDevice;
import android.view.KeyEvent;
import android.view.MotionEvent;

// Device identity and native event decoding only; modes/actions live in Go.
final class AndroidGamepad implements InputManager.InputDeviceListener {
    private final InputManager manager;
    private final Controls.Sink sink;
    private int active=-1;
    private int physical;
    AndroidGamepad(Context context,Controls.Sink sink) {
        this.sink=sink;
        manager=(InputManager)context.getSystemService(Context.INPUT_SERVICE);
        manager.registerInputDeviceListener(this,new android.os.Handler(context.getMainLooper()));
        refresh();
    }
    private static boolean qualifies(InputDevice device) {
        return device!=null && ((device.getSources() & InputDevice.SOURCE_GAMEPAD)==InputDevice.SOURCE_GAMEPAD
                || (device.getSources() & InputDevice.SOURCE_JOYSTICK)==InputDevice.SOURCE_JOYSTICK);
    }
    private void send(int kind,int a,int b) {
        if(kind==0) physical=b!=0?physical|(1<<a):physical&~(1<<a);
        if(kind==1) {
            int bit=1<<(a==4?15:16);
            if(b>=16000) physical|=bit;
            else if(b<=12000) physical&=~bit;
        }
        if(kind==3) physical=0;
        sink.send(6+kind,a,b);
    }
    void refresh() {
        if(active>=0 && qualifies(manager.getInputDevice(active))) return;
        if(active>=0) {send(3,0,0);active=-1;}
        for(int id:manager.getInputDeviceIds()) if(qualifies(manager.getInputDevice(id))) {
            active=id;send(2,0,0);sendFamily();break;
        }
    }
    void sendFamily() {
        InputDevice device=manager.getInputDevice(active);
        if(device!=null) send(4,GamepadButtons.family(device.getVendorId(),device.getName()),0);
    }
    void synchronize() {if(active>=0) {send(2,physical,0);sendFamily();}}
    boolean key(KeyEvent event) {
        InputDevice device=manager.getInputDevice(event.getDeviceId());
        if(!qualifies(device)) return false;
        int button=GamepadButtons.button(event.getKeyCode());
        if(button<0) return false;
        // Consume recognized controls from other pads, never route them as keys.
        if(event.getDeviceId()!=active || (event.getAction()==KeyEvent.ACTION_DOWN
                && event.getRepeatCount()!=0)) return true;
        if(event.getAction()==KeyEvent.ACTION_DOWN || event.getAction()==KeyEvent.ACTION_UP)
            send(0,button,event.getAction()==KeyEvent.ACTION_DOWN?1:0);
        return true;
    }
    private static int axis(InputDevice device,int preferred,int fallback) {
        if(device.getMotionRange(preferred)!=null) return preferred;
        return device.getMotionRange(fallback)!=null?fallback:-1;
    }
    private void sample(MotionEvent event,InputDevice device,int history) {
        int left=axis(device,MotionEvent.AXIS_LTRIGGER,MotionEvent.AXIS_BRAKE);
        int right=axis(device,MotionEvent.AXIS_RTRIGGER,MotionEvent.AXIS_GAS);
        if(left>=0) send(1,4,GamepadButtons.trigger(value(event,left,history)));
        if(right>=0) send(1,5,GamepadButtons.trigger(value(event,right,history)));
        if(device.getMotionRange(MotionEvent.AXIS_HAT_X)!=null) {
            float x=value(event,MotionEvent.AXIS_HAT_X,history);
            send(0,13,x<-.5f?1:0);send(0,14,x>.5f?1:0);
        }
        if(device.getMotionRange(MotionEvent.AXIS_HAT_Y)!=null) {
            float y=value(event,MotionEvent.AXIS_HAT_Y,history);
            send(0,11,y<-.5f?1:0);send(0,12,y>.5f?1:0);
        }
    }
    private static float value(MotionEvent event,int axis,int history) {
        return history<0?event.getAxisValue(axis):event.getHistoricalAxisValue(axis,history);
    }
    boolean motion(MotionEvent event) {
        InputDevice device=manager.getInputDevice(event.getDeviceId());
        if(!qualifies(device) || event.getActionMasked()!=MotionEvent.ACTION_MOVE) return false;
        if(event.getDeviceId()==active) {
            for(int h=0;h<event.getHistorySize();h++) sample(event,device,h);
            sample(event,device,-1);
        }
        return true;
    }
    public void onInputDeviceAdded(int id) {refresh();}
    public void onInputDeviceRemoved(int id) {
        if(id==active) {send(3,0,0);active=-1;}
        refresh();
    }
    public void onInputDeviceChanged(int id) {
        if(id==active) {send(3,0,0);active=-1;}
        refresh();
    }
    void close() {manager.unregisterInputDeviceListener(this);send(3,0,0);active=-1;}
}
