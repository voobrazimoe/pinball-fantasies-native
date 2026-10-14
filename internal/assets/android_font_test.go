package assets

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const androidFont = "../../hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/PixelFont.java"

// The Android overlays draw the same alphabet as the sidebar. The Java copy is
// generated from SidebarFont; regenerate it with PF_UPDATE_ANDROID_FONT=1.
func TestAndroidOverlayFontMatchesSidebarFont(t *testing.T) {
	const begin, end = "// BEGIN SidebarFont\n", "            // END SidebarFont\n"
	raw, err := os.ReadFile(androidFont)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	i, j := strings.Index(text, begin), strings.Index(text, end)
	if i < 0 || j < i {
		t.Fatal("PixelFont.java lacks the generated table markers")
	}
	var table strings.Builder
	for ch := 0; ch < 128; ch++ {
		fmt.Fprintf(&table, "            + \"%x\"", SidebarFont[ch*8:ch*8+8])
		if ch > ' ' && ch < 127 {
			fmt.Fprintf(&table, " // %c", ch)
		}
		table.WriteString("\n")
	}
	generated := text[:i+len(begin)] + table.String() + text[j:]
	if os.Getenv("PF_UPDATE_ANDROID_FONT") != "" {
		if err := os.WriteFile(androidFont, []byte(generated), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if generated != text {
		t.Fatal("PixelFont.java differs from SidebarFont; regenerate with PF_UPDATE_ANDROID_FONT=1")
	}
}
