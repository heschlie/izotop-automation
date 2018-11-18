package izotope_auto

import "C"
import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"os/exec"
	"path/filepath"
)

var inFiles = filepath.FromSlash("/Users/matthewvoelker/Desktop/to_rx/")

func main() {
	bitmaps := openBitmaps()

	// Run iZotope.
	iz := exec.Command("pathToIZotope")
	err := iz.Start()
	if err != nil {
		panic("Failed to start iZotope")
	}
	defer iz.Process.Kill()

	robotgo.ActiveName("iZotope RX 7")

	// Opens batch window.
	robotgo.KeyTap("command", "b")

	// Find the preset button.
	robotgo.BitmapClick(bitmaps["preset.bmp"])
	robotgo.Sleep(0.2)

	// Find add files button and click it.
	robotgo.BitmapClick(bitmaps["add_files.bmp"])
	robotgo.Sleep(0.2)

	// Opens text window for file path.
	robotgo.KeyTap("command", "shift", "g")
	robotgo.Sleep(0.2)
	robotgo.KeyTap("escape")
	robotgo.Sleep(0.2)
	robotgo.WriteAll(inFiles)
	robotgo.KeyTap("enter")
	robotgo.Sleep(0.2)
	robotgo.KeyTap("command", "a")
	robotgo.Sleep(0.2)
	robotgo.KeyTap("enter")

	robotgo.Sleep(10)

	// Find and click the process button.
	robotgo.BitmapClick(bitmaps["process.bmp"])

	// Capture the screen and see if we can find the cancel button, continue as long as we see it.
	screen := robotgo.CaptureScreen()
	defer robotgo.FreeBitmap(screen)

	for robotgo.CountBitmap(bitmaps["cancel.bmp"], screen) > 0 {
		robotgo.Sleep(5)
		robotgo.FreeBitmap(screen)
		screen = robotgo.CaptureScreen()
	}

	freeBitmaps(bitmaps)

	fmt.Println("Finished")
}

func openBitmaps() map[string]C.MMBitmapRef {
	names := []string{"add_files.bmp", "preset.bmp", "process.bmp"}
	bitmaps := map[string]C.MMBitmapRef{}
	for _, name := range names {
		f := robotgo.OpenBitmap(filepath.Join("images", name))
		bitmaps[name] = f
	}

	return bitmaps
}

func freeBitmaps(bitmaps map[string]C.MMBitmapRef) {
	for _, bmp := range bitmaps {
		robotgo.FreeBitmap(bmp)
	}
}
