package io.github.voobrazimoe.pinballfantasies;
import java.util.Arrays;
import java.util.ArrayList;
import java.util.List;
public final class SemanticUiTest {
    static void check(boolean value) { if(!value) throw new AssertionError(); }
    public static void main(String[] args) {
        SemanticUi ui=new SemanticUi();
        String[][] expected={{"Tap to continue"},SemanticUi.TABLES,SemanticUi.TABLES,
                {"Up","Down","Select","Back"},{"Players","−","+","Play"},{},
                {"Resume","Exit table"},{"Yes","No"},{},{"ABCDEFGHIJKLMNOPQRSTUVWXYZ"},{},{}};
        for(int mode=0;mode<12;mode++) {
            check(ui.update(mode,2,4)); check(!ui.update(mode,2,4));
            check(Arrays.equals(ui.actions(),expected[mode]));
            for(String utility:ui.utilities()) check(!utility.contains("Data") && !utility.contains("Import"));
            check(ui.gameplay()==(mode==5) && ui.plunger()==(mode==5));
            for(String action:ui.actions()) check(!action.matches("F[1-8]|Enter|Esc|P|M|Y|N"));
        }
        for(int mode=0;mode<12;mode++) {
            ui.update(mode,1,8);check(ui.controllerHints()==(mode!=SemanticUi.INITIALS));
            ui.update(mode,1,0);check(!ui.controllerHints());
        }
        for(int i=0;i<5;i++) check(ui.code(SemanticUi.TABLES[i])==59+i);
        ui.update(4,1,0);
        List<Integer> makes=new ArrayList<>();
        Controls c=new Controls((k,a,b)->{check(k==1); makes.add(a);},(ms,r)->{}); c.enabled=true;
        for(int players=1;players<=8;players++) {
            check(ui.players==players); c.tap(ui.code("Play")); check(makes.get(players-1)==58+players);
            ui.players(1);
        }
        check(makes.size()==8 && ui.players==8); ui.players(-99); check(ui.players==1);
        ui.players(5); ui.update(5,1,0); ui.update(4,1,0); check(ui.players==1);
        check(ui.code("Resume")==28 && ui.code("Exit table")==1);
        check(ui.code("Yes")==21 && ui.code("No")==49);
        check(ui.code("Up")==72 && ui.code("Down")==80 && ui.code("Select")==28 && ui.code("Back")==1);
        String letters="ABCDEFGHIJKLMNOPQRSTUVWXYZ";
        int[] codes={30,48,46,32,18,33,34,35,23,36,37,38,50,49,24,25,16,19,31,20,22,47,17,45,21,44};
        for(int i=0;i<26;i++) check(ui.code(letters.substring(i,i+1))==codes[i]);
        check(ui.code("1")==-1 && ui.code("0")==-1);
        List<String> events=new ArrayList<>();
        Controls keyboard=new Controls((k,a,b)->events.add(k+":"+a+":"+b),(ms,r)->{});
        keyboard.enabled=true;
        keyboard.key(59,true,0); check(keyboard.held[0]);
        keyboard.keyboardMode(true);
        check(keyboard.keyboardMode && !keyboard.held[0]);
        events.clear(); keyboard.key(29,true,0); // A still reaches the DOS make path.
        check(events.contains("1:30:0"));
        keyboard.key(60,true,0); check(keyboard.held[1]);
        keyboard.keyboardMode(false); check(!keyboard.keyboardMode && !keyboard.held[1]);
        check(events.contains("0:1:0"));
        System.out.println("PASS: A6 authoritative mode mapping, semantic makes, exact Play count and initials");
    }
}
