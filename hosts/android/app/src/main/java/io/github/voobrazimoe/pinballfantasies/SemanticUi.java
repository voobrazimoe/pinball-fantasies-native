package io.github.voobrazimoe.pinballfantasies;

// One framework-free mapper. Mode numbers are the ABI's authoritative modes.
final class SemanticUi {
    static final int STARTUP=0, SELECTOR=1, SELECTOR_TEXT=2, OPTIONS=3, ATTRACT=4,
            PLAYING=5, PAUSED=6, QUESTION=7, GAME_END=8, INITIALS=9, ENTRY_WAIT=10, QUIT=11;
    static final String[] TABLES={"Party Land","Speed Devils","Billion Dollar Gameshow","Stones 'N Bones","Options"};
    int mode=-1, table, flags, players=1;
    boolean update(int mode,int table,int flags) {
        if (this.mode==mode && this.table==table && this.flags==flags) return false;
        if (mode==ATTRACT && (this.mode!=ATTRACT || this.table!=table)) players=1;
        this.mode=mode; this.table=table; this.flags=flags; return true;
    }
    boolean gameplay() { return mode==PLAYING; }
    boolean plunger() { return gameplay() && (flags&4)!=0; }
    void players(int delta) { players=Math.max(1,Math.min(8,players+delta)); }
    String[] actions() {
        switch(mode) {
            case STARTUP: return new String[]{"Tap to continue"};
            case SELECTOR: case SELECTOR_TEXT: return TABLES.clone();
            case ATTRACT: return new String[]{"Players","−","+","Play"};
            case OPTIONS: return new String[]{"Up","Down","Select","Back"};
            case PAUSED: return new String[]{"Resume","Exit table"};
            case QUESTION: return new String[]{"Yes","No"};
            case INITIALS: return new String[]{"ABCDEFGHIJKLMNOPQRSTUVWXYZ"};
            default: return new String[0];
        }
    }
    String[] utilities() {
        switch(mode) {
            case PLAYING: return new String[]{"Pause","Music","Data / Import DOS folder","Advanced keyboard"};
            case ATTRACT: return new String[]{"Back to tables","Data / Import DOS folder","Advanced keyboard"};
            default: return new String[]{"Data / Import DOS folder","Advanced keyboard"};
        }
    }
    int code(String action) {
        for(int i=0;i<TABLES.length;i++) if(TABLES[i].equals(action)) return 59+i;
        switch(action) {
            case "Tap to continue": return 57;
            case "Play": return 58+players;
            case "Up": return 72; case "Down": return 80;
            case "Select": case "Resume": return 28;
            case "Back": case "Exit table": case "Back to tables": return 1;
            case "Pause": return 25; case "Music": return 50;
            case "Yes": return 21; case "No": return 49;
            default: return action.length()==1 && action.charAt(0)>='A' && action.charAt(0)<='Z'
                    ? Controls.make(29+action.charAt(0)-'A') : -1;
        }
    }
}
