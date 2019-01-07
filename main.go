package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-vgo/robotgo"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Filepaths from dropbox.
var srcDir_rs = "/Users/matthewvoelker/Documents/"
var srcDir_nv = "/Users/matthewvoelker/Documents/"

// Where to point izotope.
var inFiles_rs = filepath.FromSlash("/Users/matthewvoelker/Desktop/to_rx/")
var inFiles_nv = filepath.FromSlash("/Users/matthewvoelker/Desktop/to_rx/")

// Locations for RS talents.
var stage_rs = "Test_RS_00_Dailies"
var roughPath_rs = "Test_RS_01_Rough_Edits"
var finishedPath_rs = "Test_RS_02_RX_Files"

// Locations for NV talents
var stage_nv = "Test_NV_00_Dailies"
var roughPath_nv = "Test_NV_01_Rough_Edits"
var finishedPath_nv = "Test_NV_02_RX_Files"

var AudioFiles_rs []AudioFile
var AudioFiles_nv []AudioFile

type AudioFile struct {
	Name      string
	SrcPath   string
	RoughPath string
	RXPath    string
}

func main() {
	AudioFiles_rs = []AudioFile{}
	AudioFiles_nv = []AudioFile{}
	findNewFiles(srcDir_rs, stage_rs, roughPath_rs, finishedPath_rs, AudioFiles_rs)
	findNewFiles(srcDir_nv, stage_nv, roughPath_nv, finishedPath_nv, AudioFiles_nv)
	moveFilesForProcessing(inFiles_rs, AudioFiles_rs)
	moveFilesForProcessing(inFiles_nv, AudioFiles_nv)
	runIZotope()
	moveFinishedFiles(inFiles_rs, AudioFiles_rs)
	moveFinishedFiles(inFiles_nv, AudioFiles_nv)
}

// Walk the relevant part of the filesystem to find the files to be processed.
func findNewFiles(srcDir, stage, roughPath, finishedPath string, AudioFiles []AudioFile) {
	err := filepath.Walk(srcDir+stage,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories starting with "_".
			if info.IsDir() && strings.HasPrefix(info.Name(), "_") {
				return filepath.SkipDir
			}

			// Skip any non wav files.
			if !strings.HasSuffix(info.Name(), ".wav") {
				return nil
			}

			// Generate our paths and store this in a slice to be processed.
			rPath := strings.Replace(path, stage, roughPath, 1)
			fPath := strings.Replace(path, stage, finishedPath, 1)
			aFile := AudioFile{
				Name:      info.Name(),
				SrcPath:   path,
				RoughPath: rPath,
				RXPath:    fPath,
			}
			AudioFiles = append(AudioFiles, aFile)
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	aFilesJson, _ := json.Marshal(AudioFiles)
	name := "logs.json"
	if strings.Contains(stage, "RS") {
		name = "rs-logs.json"
	} else if strings.Contains(stage, "NV") {
		name = "nv-logs.json"
	}
	err = ioutil.WriteFile(name, aFilesJson, 0644)
	if err != nil {
		fmt.Println("Failed to write log file to logs.json")
	}
}

// Move all of our files into the applications working directory.
func moveFilesForProcessing(inFiles string, AudioFiles []AudioFile) {
	for _, file := range AudioFiles {
		copy(file.SrcPath, filepath.Join(inFiles, file.Name))
		os.Remove(file.SrcPath)
	}
}

// moveFinishedFiles will move the rough file and the finished files to their respective places and
// delete them from the applications working directory.
func moveFinishedFiles(inFiles string, AudioFiles []AudioFile) {
	files, err := ioutil.ReadDir(inFiles)
	if err != nil {
		panic(fmt.Sprintf("Failed to read files from %s", inFiles))
	}
	for _, file := range files {
		fName := filepath.Join(inFiles, file.Name())
		for _, afile := range AudioFiles {
			// Copy the rough file over and remove it.
			if file.Name() == afile.Name {
				copy(fName, afile.RoughPath)
				os.Remove(fName)
			// Copy the finished file over and remove it.
			} else if strings.HasPrefix(file.Name(), strings.TrimSuffix(afile.Name, ".wav")) {
				copy(fName, afile.RXPath)
				os.Remove(fName)
			}
		}
	}
}

func copy(src, dst string) (int64, error) {
	return 0, nil
}

func runIZotope() {
	// Run iZotope.
	iz := exec.Command("pathToIZotope")
	err := iz.Start()
	if err != nil {
		panic("Failed to start iZotope")
	}
	defer iz.Process.Kill()

	robotgo.ActiveName("iZotope RX 7")

	// Opens batch window.
	robotgo.KeyTap("b", "command")

	// Find the preset button.
	bmp := robotgo.OpenBitmap("images/preset.bmp")
	x, y := robotgo.FindBitmap(bmp)
	fmt.Println(x)
	fmt.Println(y)
	robotgo.MoveClick(x, y)
	robotgo.Sleep(0.2)

	// Find add files button and click it.
	robotgo.BitmapClick(robotgo.OpenBitmap("images/add_files.bmp"))
	robotgo.Sleep(0.2)

	// Opens text window for file path.
	robotgo.KeyTap("g", "shift", "command")
	robotgo.Sleep(0.2)
	robotgo.KeyTap("escape")
	robotgo.Sleep(0.2)
	robotgo.WriteAll(inFiles)
	robotgo.KeyTap("enter")
	robotgo.Sleep(0.2)
	robotgo.KeyTap("a", "command")
	robotgo.Sleep(0.2)
	robotgo.KeyTap("enter")

	robotgo.Sleep(10)

	// Find and click the process button.
	robotgo.BitmapClick(robotgo.OpenBitmap("images/process.bmp"))

	// Capture the screen and see if we can find the cancel button, continue as long as we see it.
	screen := robotgo.CaptureScreen()
	defer robotgo.FreeBitmap(screen)

	for robotgo.CountBitmap(robotgo.OpenBitmap("images/cancel.bmp"), screen) > 0 {
		robotgo.Sleep(5)
		robotgo.FreeBitmap(screen)
		screen = robotgo.CaptureScreen()
	}

	fmt.Println("Finished")
}
