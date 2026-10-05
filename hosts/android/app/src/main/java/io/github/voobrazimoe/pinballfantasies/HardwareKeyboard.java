package io.github.voobrazimoe.pinballfantasies;

import android.content.Context;
import android.hardware.input.InputManager;
import android.view.InputDevice;

// One device inventory owns the policy. Configuration changes trigger a rescan;
// Configuration.keyboard alone cannot distinguish virtual or built-in devices.
final class HardwareKeyboard implements InputManager.InputDeviceListener {
    interface Changed { void accept(boolean present); }
    private final InputManager manager;
    private final Changed changed;
    HardwareKeyboard(Context context,Changed changed) {
        this.changed=changed;
        manager=(InputManager)context.getSystemService(Context.INPUT_SERVICE);
        manager.registerInputDeviceListener(this,new android.os.Handler(context.getMainLooper()));
        refresh();
    }
    static boolean qualifies(boolean external,boolean virtual,int type,int sources) {
        return external && !virtual && type==InputDevice.KEYBOARD_TYPE_ALPHABETIC
                && (sources&InputDevice.SOURCE_KEYBOARD)==InputDevice.SOURCE_KEYBOARD;
    }
    void refresh() {
        boolean present=false;
        for(int id:manager.getInputDeviceIds()) {
            InputDevice device=manager.getInputDevice(id);
            if(device!=null && qualifies(device.isExternal(),device.isVirtual(),device.getKeyboardType(),device.getSources())) {
                present=true; break;
            }
        }
        changed.accept(present);
    }
    public void onInputDeviceAdded(int id) { refresh(); }
    public void onInputDeviceChanged(int id) { refresh(); }
    public void onInputDeviceRemoved(int id) { refresh(); }
    void close() { manager.unregisterInputDeviceListener(this); }
}
