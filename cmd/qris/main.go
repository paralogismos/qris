// main.go
//
// CLI interface using the qris library.
//
// Maybe a simple CLI interface would take a path argument that provides
// the path to a file containing a list of files to parse.
//
// Should the list of files contain complete paths, or filenames only?

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"qris"
)

func main() {
	// Parse command line flags
	conf := flag.Bool("config", false, "Show path to configuration file.")
	dir := flag.String("dir", "",
		"Set the current working directory.")
	filePath := flag.String("f", "",
		"Path to a file to be parsed, relative to working directory.")
	batchPath := flag.String("b", "",
		"Path to a directory containing files to be parsed, relative to working directory.")
	valid := flag.Bool("valid", false, "Validate UTF8 files.")
	flag.Parse()

	if *filePath != "" && *batchPath != "" {
		fmt.Println("-f and -b flags may not be used together")
		flag.Usage()
		os.Exit(1)
	}

	// Check for config directory and create if missing.
	const configDir = "qris"
	const configFile = "qris.conf"
	var configPath string
	userConfig, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Unable to access user configuration directory")
	} else {
		configDirPath := filepath.Join(userConfig, configDir)
		_, err := os.ReadDir(configDirPath)
		if err != nil {
			err := os.Mkdir(configDirPath, 0666)
			if err != nil {
				fmt.Println("Unable to create configuration directory")
			}
		}
		configPath = filepath.Join(userConfig, configDir, configFile)
	}

	if *conf {
		fmt.Println("Configuration file at", configPath)
	}

	// Get current working directory.
	// Look in home directory (~/qris/) for a config file that stores
	// a working directory path. If it exists, set the current working
	// directory accordingly. Otherwise use `os.Getwd()` to get it from
	// the system.
	var workDir string
	config, err := os.Open(configPath)
	if err == nil {
		scanner := bufio.NewScanner(config)
		if scanner.Scan() {
			workDir = scanner.Text()
		}
	} else {
		workDir, err = os.Getwd()
		if err != nil {
			fmt.Println("Unable to get current working directory")
			os.Exit(1)
		}
	}
	config.Close()

	// Set working directory.
	// Store the new current working directory path in a configuration file.
	if *dir != "" {
		workDir, err = filepath.Abs(*dir)
		if err != nil {
			fmt.Println("Unable to create new working directory path")
			os.Exit(1)
		}
		if os.Chdir(workDir) != nil {
			fmt.Println("Unable to update working directory")
			os.Exit(1)
		}
		config, err := os.Create(configPath)
		if err != nil {
			fmt.Println("Unable to create configuration file")
		} else {
			fmt.Println("Storing config data:", workDir)
			fmt.Fprintln(config, workDir)
		}
		config.Close()
	}

	// Always show current working directory.
	fmt.Println("Working in directory", workDir)

	// `dataList` is a list of files to be processed.
	var dataList []string

	// `workPath` is the absolute path to files to be processed.
	// `batchPath` may include directory structure relative to the
	// working directory, and this additional directory structure is
	// included in `workPath`.
	var workPath string

	// First populate `dataList` with any batch files.
	if *batchPath != "" {
		// Allow dot argument to indicate batch files found in working directory.
		if workPath == "." {
			workPath = workDir
		} else {
			workPath, _ = filepath.Abs(*batchPath)
		}
		files, err := os.ReadDir(workPath)
		if err != nil {
			log.Fatal(err)
		}
		// create list of files in working directory
		for _, file := range files {
			dataList = append(dataList, file.Name())
		}
	} else {
		// Otherwise add a single file to `dataList` if one was supplied.
		if *filePath != "" {
			workPath, err = filepath.Abs(*filePath)
			var workFile string
			workPath, workFile = filepath.Split(workPath)
			dataList = append(dataList, workFile)
		}
	}

	if *valid {
		// Only validate files.
		n := len(dataList)
		if n == 1 {
			fmt.Println("Validating one file....")
		} else {
			fmt.Printf("Validating %d files....\n", n)
		}
		allPassed := true
		for _, file := range dataList {
			fmt.Println(file)
			vFile := filepath.Join(workPath, file)
			isPassed := qris.ValidateUTF8(vFile)
			allPassed = allPassed && isPassed
		}
		if allPassed {
			fmt.Println("All files were valid UTF8.")
		}
	} else {
		// Parse all files.
		for _, file := range dataList {
			// File to store parsed quotes
			pFile := filepath.Join(workPath, file)
			base := strings.TrimSuffix(pFile, filepath.Ext(pFile))
			pQuotes := base + "_PARSED.ris"

			// File to store discarded lines
			pDiscard := base + "_DISCARD.txt"

			pf := qris.ParseFile(pFile)
			qris.WriteDiscards(pf.Discards, pDiscard)
			qris.WriteQuotes(&pf, pQuotes)
		}
	}
}
