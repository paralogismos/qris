// configure.go
//
// Configuration definitions and functions.
package qris

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Definitions of system constants.
const Version = "v0.18.1"
const parsedSuffix = "_PARSED.ris"
const discardSuffix = "_DISCARD.txt"
const configDir = "qris"
const configFile = "qris.conf"

// Set platform-specific line ending.
var LineEnding string = PlatformLineEnding()

func PlatformLineEnding() string {
	var lineEnding string
	switch runtime.GOOS {
	case "windows":
		lineEnding = "\r\n"
	default:
		lineEnding = "\n"
	}
	return lineEnding
}

// `GetConfigPath` checks for a configuration directory and
// creates one if none exists.
func GetConfigPath() string {
	configPath := ""
	userConfig, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		configDirPath := filepath.Join(userConfig, configDir)
		_, err := os.ReadDir(configDirPath)
		if err != nil {
			err := os.Mkdir(configDirPath, 0666)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		configPath = filepath.Join(userConfig, configDir, configFile)
	}
	return configPath
}

// `GetWorkDir` looks in `configPath` for a configuration file. If one exists,
// the configured working directory is returned. Otherwise `os.Getwd()` is used
// to get the current working directory from the system and this path is returned.
func GetWorkDir(configPath string) string {
	workDir := ""
	config, err := os.Open(configPath)
	if err == nil {
		defer config.Close()
		scanner := bufio.NewScanner(config)
		if scanner.Scan() {
			workDir = scanner.Text()
			if err := os.Chdir(workDir); err != nil {
				fmt.Fprintln(os.Stderr, err)
				workDir, err = os.Getwd()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}
	} else {
		workDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	return workDir
}

// `SetWorkDir` stores a new working directory path in the configuration file
// and sets the current working directory on the users system.
// If `SetWorkDir` cannot set the working directory for some reason, the
// default working directory already established by the system and discovered
// by `GetWorkDir` should be available for the user to use.
func SetWorkDir(dirPath, configPath string) {
	if dirPath != "" {
		workDir, err := filepath.Abs(dirPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.Chdir(workDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		config, err := os.Create(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			defer config.Close()
			fmt.Fprintln(config, workDir)
		}
	}
}

// *** I don't think that this function is being used anymore....
// *** Maybe this should be removed.
// `GetFileList` creates a list of filenames from the text file specified by `fpath`.
// This file should have one filename per line.
func GetFileList(fpath string) []string {
	file, err := os.Open(fpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	files := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}

	return files
}

// `GetBatchList` takes a path argument and returns a list of all files found
// in the directory specified by the path. Directories found in the specified
// directory are not included in the list. It is an error if the path does not
// lead to a directory.
func GetBatchList(path string) []string {
	var dataList []string
	dirEnts, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Create list of files in specified directory.
	for _, dirEnt := range dirEnts {
		if dirEnt.IsDir() {
			continue
		}
		dataList = append(dataList, dirEnt.Name())
	}

	return dataList
}

// `GetWorkPath` takes as arguments an absolute path to the current working
// directory, a path to a batch directory to be processed (`bFlag`), and a path
// to a file to be processed (`fFlag`). The second two paths are relative to
// the current working directory. This information is used to create a list
// of files for processing which is returned to the caller.
func GetWorkPath(workDir, bFlag, fFlag string) ([]string, string) {
	var dList []string
	var wPath string
	if bFlag == "" {
		if fFlag != "" {
			// Add a single file to `dataList` if one was supplied.
			wPath, _ = filepath.Abs(fFlag)
			var wFile string
			wPath, wFile = filepath.Split(wPath)
			dList = append(dList, wFile)
		}
	} else {
		// Batch process files.
		// Allow dot argument to indicate batch files found in working directory.
		if bFlag == "." {
			wPath = workDir
		} else {
			wPath, _ = filepath.Abs(bFlag)
		}
		dList = GetBatchList(wPath)
	}
	return dList, wPath
}
