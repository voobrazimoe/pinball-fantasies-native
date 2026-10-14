package io.github.voobrazimoe.pinballfantasies;

import android.graphics.Canvas;
import android.graphics.Paint;
import java.util.ArrayList;
import java.util.List;

/** The selector sidebar's 8x8 alphabet, drawn at whole-pixel scales by the overlays. */
final class PixelFont {
    static final int CELL=8;
    static final char UP=0x18, DOWN=0x19, RIGHT=0x1a, LEFT=0x1b, RAISE=0x1e, LOWER=0x1f;
    private static final String CELLS = ""
            // BEGIN SidebarFont
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "0000000000000000"
            + "3078fc3030303000"
            + "30303030fc783000"
            + "00180cfe0c180000"
            + "0060c0fec0600000"
            + "0000000000000000"
            + "0000000000000000"
            + "00003078fc000000"
            + "0000fc7830000000"
            + "0000000000000000"
            + "3030303030003000" // !
            + "0000000000000000" // "
            + "0000000000000000" // #
            + "0000000000000000" // $
            + "ccd81830606ccc00" // %
            + "0000000000000000" // &
            + "3030600000000000" // '
            + "1830606060301800" // (
            + "6030181818306000" // )
            + "00cc78fc78cc0000" // *
            + "003030fc30300000" // +
            + "0000000000306000" // ,
            + "000000fc00000000" // -
            + "0000000000303000" // .
            + "0c0c183060c0c000" // /
            + "78ccdcfceccc7800" // 0
            + "307030303030fc00" // 1
            + "78cc0c3860c0fc00" // 2
            + "78cc0c380ccc7800" // 3
            + "1c3c6cccfc0c0c00" // 4
            + "fcc0f80c0ccc7800" // 5
            + "78c0c0f8cccc7800" // 6
            + "fc0c183030303000" // 7
            + "78cccc78cccc7800" // 8
            + "78cccc7c0c0c7800" // 9
            + "0030300030300000" // :
            + "0000000000000000" // ;
            + "183060c060301800" // <
            + "0000fc00fc000000" // =
            + "6030180c18306000" // >
            + "78cc0c3830003000" // ?
            + "0000000000000000" // @
            + "78ccccfccccccc00" // A
            + "f8ccccf8ccccf800" // B
            + "78ccc0c0c0cc7800" // C
            + "f8ccccccccccf800" // D
            + "fcc0c0f8c0c0fc00" // E
            + "fcc0c0f8c0c0c000" // F
            + "78ccc0dccccc7c00" // G
            + "ccccccfccccccc00" // H
            + "fc3030303030fc00" // I
            + "3c0c0c0ccccc7800" // J
            + "ccd8f0e0f0d8cc00" // K
            + "c0c0c0c0c0c0fc00" // L
            + "c6eefed6c6c6c600" // M
            + "c6e6f6decec6c600" // N
            + "78cccccccccc7800" // O
            + "f8ccccf8c0c0c000" // P
            + "78ccccccdc780c00" // Q
            + "f8ccccf8f0d8cc00" // R
            + "78ccc0780ccc7800" // S
            + "fc30303030303000" // T
            + "cccccccccccc7800" // U
            + "cccccccccc783000" // V
            + "c6c6c6d6feeec600" // W
            + "cccc783078cccc00" // X
            + "cccccc7830303000" // Y
            + "fc0c183060c0fc00" // Z
            + "0000000000000000" // [
            + "0000000000000000" // \
            + "0000000000000000" // ]
            + "0000000000000000" // ^
            + "000000000000fc00" // _
            + "0000000000000000" // `
            + "0000000000000000" // a
            + "0000000000000000" // b
            + "0000000000000000" // c
            + "0000000000000000" // d
            + "0000000000000000" // e
            + "0000000000000000" // f
            + "0000000000000000" // g
            + "0000000000000000" // h
            + "0000000000000000" // i
            + "0000000000000000" // j
            + "0000000000000000" // k
            + "0000000000000000" // l
            + "0000000000000000" // m
            + "0000000000000000" // n
            + "0000000000000000" // o
            + "0000000000000000" // p
            + "0000000000000000" // q
            + "0000000000000000" // r
            + "0000000000000000" // s
            + "0000000000000000" // t
            + "0000000000000000" // u
            + "0000000000000000" // v
            + "0000000000000000" // w
            + "0000000000000000" // x
            + "0000000000000000" // y
            + "0000000000000000" // z
            + "0000000000000000" // {
            + "0000000000000000" // |
            + "0000000000000000" // }
            + "0000000000000000" // ~
            + "0000000000000000"
            // END SidebarFont
            ;
    private static final byte[] BITS=new byte[128*CELL];
    static {
        for(int i=0;i<BITS.length;i++) BITS[i]=(byte)Integer.parseInt(CELLS.substring(i*2,i*2+2),16);
    }
    private PixelFont() {}

    /** Maps display text onto the font: capitals, the overlay arrows, '?' for the rest. */
    static String normalize(String text) {
        StringBuilder out=new StringBuilder(text.length());
        for(int i=0;i<text.length();i++) {
            char c=Character.toUpperCase(text.charAt(i));
            switch(c) {
                case '↑': c=UP; break; case '↓': c=DOWN; break;
                case '→': c=RIGHT; break; case '←': c=LEFT; break;
                case '▲': c=RAISE; break; case '▼': c=LOWER; break;
                case '−': case '–': case '—': c='-'; break;
                case '’': c='\''; break;
                case '…': out.append(".."); c='.'; break;
                default: break;
            }
            if(c>=128 || (c!=' ' && !drawable(c))) c='?';
            out.append(c);
        }
        return out.toString();
    }
    static boolean drawable(char c) {
        if(c>=128) return false;
        for(int row=0;row<CELL;row++) if(BITS[c*CELL+row]!=0) return true;
        return false;
    }
    static int width(String text,int scale) { return text.length()*CELL*scale; }

    /** Draws normalized text with its cell's top-left corner at (x, y). */
    static void draw(Canvas canvas,String text,float x,float y,int scale,Paint paint) {
        for(int i=0;i<text.length();i++) {
            char c=text.charAt(i);
            if(c>=128) continue;
            float left=x+i*CELL*scale;
            for(int row=0;row<CELL;row++) {
                int bits=BITS[c*CELL+row]&255;
                int col=0;
                while(col<CELL) {
                    if((bits&(128>>col))==0) { col++; continue; }
                    int run=col;
                    while(run<CELL && (bits&(128>>run))!=0) run++;
                    canvas.drawRect(left+col*scale,y+row*scale,left+run*scale,y+(row+1)*scale,paint);
                    col=run;
                }
            }
        }
    }

    /** Greedy word wrap at whole characters; long words are split. */
    static List<String> wrap(String text,int columns) {
        List<String> lines=new ArrayList<>();
        columns=Math.max(1,columns);
        StringBuilder line=new StringBuilder();
        for(String word:text.trim().split(" +")) {
            while(word.length()>columns) {
                if(line.length()>0) { lines.add(line.toString()); line.setLength(0); }
                lines.add(word.substring(0,columns)); word=word.substring(columns);
            }
            if(line.length()>0 && line.length()+1+word.length()>columns) { lines.add(line.toString()); line.setLength(0); }
            if(line.length()>0) line.append(' ');
            line.append(word);
        }
        if(line.length()>0 || lines.isEmpty()) lines.add(line.toString());
        return lines;
    }

    /** Largest whole scale whose cell fits the requested pixel height, never below one. */
    static int scale(float pixels) { return Math.max(1,(int)Math.floor(pixels/CELL)); }
}
