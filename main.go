package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Filepaths from dropbox.
var srcDir_rs = filepath.FromSlash("/Users/matthewvoelker/Documents/Editorial/RS")
var srcDir_nv = filepath.FromSlash("/Users/matthewvoelker/Documents/Editorial/NV")

// Where to point izotope.
var inFiles_rs = filepath.FromSlash("/Users/matthewvoelker/Desktop/RS_to_RX")
var inFiles_nv = filepath.FromSlash("/Users/matthewvoelker/Desktop/NV_to_RX")

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
	findNewFiles(srcDir_rs, stage_rs, roughPath_rs, finishedPath_rs, &AudioFiles_rs)
	findNewFiles(srcDir_nv, stage_nv, roughPath_nv, finishedPath_nv, &AudioFiles_nv)
	moveFilesForProcessing(inFiles_rs, AudioFiles_rs)
	moveFilesForProcessing(inFiles_nv, AudioFiles_nv)
	//runIZotope()
	moveFinishedFiles(inFiles_rs, AudioFiles_rs)
	moveFinishedFiles(inFiles_nv, AudioFiles_nv)
}

// Walk the relevant part of the filesystem to find the files to be processed.
func findNewFiles(srcDir, stage, roughPath, finishedPath string, AudioFiles *[]AudioFile) {
	err := filepath.Walk(filepath.Join(srcDir, stage),
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
			fmt.Printf("Created %v\n", aFile)
			*AudioFiles = append(*AudioFiles, aFile)
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	aFilesJson, _ := json.MarshalIndent(AudioFiles, "", "  ")
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
		fmt.Println("Copying " + file.SrcPath)
		_, err := copy(file.SrcPath, filepath.Join(inFiles, file.Name))
		if err != nil {
			fmt.Println(err)
		}
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
				fmt.Println("Copying " + fName + "to " + afile.RoughPath)
				_, err := copy(fName, afile.RoughPath)
				if err != nil {
					fmt.Println(err)
				}
				os.Remove(fName)
			// Copy the finished file over and remove it.
			} else if strings.HasPrefix(file.Name(), strings.TrimSuffix(afile.Name, ".wav")) {
				fmt.Println("Copying " + fName + "to " + afile.RXPath)
				_, err := copy(fName, afile.RXPath)
				if err != nil {
					fmt.Println(err)
				}
				os.Remove(fName)
			}
		}
	}
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func runIZotope() {
}
