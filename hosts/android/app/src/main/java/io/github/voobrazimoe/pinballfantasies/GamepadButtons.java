package io.github.voobrazimoe.pinballfantasies;

// Framework-free Android keycode to standard gamepad button contract.
final class GamepadButtons {
    static int family(int vendor,String name) {
        String lower=name==null?"":name.toLowerCase(java.util.Locale.ROOT);
        if(vendor==0x054c || lower.contains("dualshock") || lower.contains("dualsense") || lower.contains("playstation")) return 2;
        if(vendor==0x045e || lower.contains("xbox")) return 1;
        if(vendor==0x057e) return 3;
        return 0;
    }
    static int button(int key) {
        switch(key) {
            case 96: return 0; case 97: return 1; case 99: return 2; case 100: return 3;
            case 109: return 4; case 110: return 5; case 108: return 6;
            case 106: return 15; case 107: return 16;
            case 102: return 7; case 103: return 8; case 104: return 9; case 105: return 10;
            case 19: return 11; case 20: return 12; case 21: return 13; case 22: return 14;
            default: return -1;
        }
    }
    static int trigger(float value) {
        return Math.round(Math.max(0,Math.min(1,value))*32767);
    }
}
