package io.github.voobrazimoe.pinballfantasies;

public final class GamepadButtonsTest {
    public static void main(String[] args) {
        if(GamepadButtons.family(0x054c,"Wireless Controller")!=2
                || GamepadButtons.family(0,"DualSense Wireless")!=2
                || GamepadButtons.family(0x045e,"Gamepad")!=1
                || GamepadButtons.family(0,"Xbox Wireless")!=1
                || GamepadButtons.family(0x057e,"Pro Controller")!=3
                || GamepadButtons.family(0,"Unknown")!=0) throw new AssertionError("controller family");
        int[][] mappings={{96,0},{97,1},{99,2},{100,3},{109,4},{110,5},{108,6},
                {102,7},{103,8},{104,9},{105,10},{19,11},{20,12},{21,13},{22,14},{106,15},{107,16}};
        for(int[] mapping:mappings) if(GamepadButtons.button(mapping[0])!=mapping[1])
            throw new AssertionError("Android button identity: "+mapping[0]);
        for(int key:new int[]{4,29,62,66,111,131,160}) if(GamepadButtons.button(key)!=-1)
            throw new AssertionError("Keyboard key classified as gamepad: "+key);
        if(GamepadButtons.trigger(-1)!=0 || GamepadButtons.trigger(2)!=32767
                || GamepadButtons.trigger(.5f)!=16384) throw new AssertionError("trigger normalization");
        System.out.println("PASS: Android gamepad button identities and normalized triggers");
    }
}
