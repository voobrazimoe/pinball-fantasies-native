//go:build linux

package platform

/*
#include <SDL.h>
#include <stdlib.h>
static int pf_launch_choice(const char *text) {
 const SDL_MessageBoxButtonData buttons[] = {
  {SDL_MESSAGEBOX_BUTTON_ESCAPEKEY_DEFAULT, 2, "Quit"},
  {0, 1, "Play demo"},
  {SDL_MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, 0, "Import game..."},
 };
 const SDL_MessageBoxData data = {SDL_MESSAGEBOX_INFORMATION, NULL, "Pinball Fantasies", text, 3, buttons, NULL};
 int choice = 2;
 if (SDL_ShowMessageBox(&data, &choice) != 0) return -1;
 return choice;
}
static void pf_launch_message(const char *text) {
 SDL_ShowSimpleMessageBox(SDL_MESSAGEBOX_WARNING, "Pinball Fantasies", text, NULL);
}
*/
import "C"
import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unsafe"
)

func askLaunch() launchChoice {
	text := C.CString(launchText)
	defer C.free(unsafe.Pointer(text))
	switch C.pf_launch_choice(text) {
	case 0:
		return launchImport
	case 1:
		return launchDemo
	case -1:
		// No display for a dialog: keep the old command-line behaviour.
		fmt.Fprintln(os.Stderr, "No original game files: place them beside the program or use -data-dir <folder>.")
	}
	return launchQuit
}

// The desktop's own folder chooser runs as a separate program, so the AppImage
// needs no toolkit of its own.
func pickFolder() (string, bool) {
	title := "Choose the original Pinball Fantasies folder (with INTRO.PRG and TABLE1-4)"
	for _, command := range [][]string{
		{"zenity", "--file-selection", "--directory", "--title=" + title},
		{"kdialog", "--getexistingdirectory", os.Getenv("HOME"), "--title", title},
	} {
		if _, err := exec.LookPath(command[0]); err != nil {
			continue
		}
		out, err := exec.Command(command[0], command[1:]...).Output()
		if err != nil {
			return "", false
		}
		path := strings.TrimRight(string(out), "\n")
		return path, path != ""
	}
	showMessage("No folder chooser (zenity or kdialog) was found.\n\nCopy the original game files next to the AppImage, or start it with -data-dir <folder>.")
	return "", false
}

func showMessage(text string) {
	c := C.CString(text)
	defer C.free(unsafe.Pointer(c))
	C.pf_launch_message(c)
}
