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
	//	"bufio"
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
	dir := flag.String("d", "",
		"Set the current working directory.")
	filePath := flag.String("f", "",
		"Path to a file to be parsed, absolute or relative.")
	batchPath := flag.String("b", "",
		"Path to a directory containing files to be parsed, absolute or relative.")
	validate := flag.Bool("v", false, "Validate UTF8 files.")
	flag.Parse()

	if *filePath != "" && *batchPath != "" {
		fmt.Fprintln(os.Stderr, "-f and -b flags may not be used together")
		flag.Usage()
		os.Exit(1)
	}

	// Configure the system.
	configPath := qris.GetConfigPath()
	if *conf {
		fmt.Println("Configuration file at", configPath)
	}

	// Set working directory.
	// Store the new current working directory path in a configuration file.
	if *dir != "" {
		workDir, err := filepath.Abs(*dir)
		if err != nil {
			fmt.Fprintln(os.Stderr,
				"Unable to create new working directory path")
			os.Exit(1)
		}
		if os.Chdir(workDir) != nil {
			fmt.Fprintln(os.Stderr,
				"Unable to update working directory")
			os.Exit(1)
		}
		config, err := os.Create(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr,
				"Unable to create configuration file")
		} else {
			fmt.Fprintln(config, workDir)
		}
		config.Close()
	}

	// Get current working directory.
	workDir := qris.GetWorkDir(configPath)

	// Always show current qris version
	fmt.Println("qris version", qris.Version)

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
		if *batchPath == "." {
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
			workPath, _ := filepath.Abs(*filePath)
			var workFile string
			workPath, workFile = filepath.Split(workPath)
			dataList = append(dataList, workFile)
		}
	}

	// Parse all files.
	allPassed := true // For UTF8 validation option
	for _, file := range dataList {
		// Skip any file not ending with .txt extension.
		// Note that directories ending with .txt WILL be
		// processed, and this will cause a panic.
		if filepath.Ext(file) != ".txt" {
			continue
		}

		// Display file name as it is processed
		fmt.Println(file)

		// File path to process
		pFile := filepath.Join(workPath, file)

		if *validate {
			allPassed = allPassed && qris.ValidateUTF8(pFile)
		}

		// File to store parsed quotes
		base := strings.TrimSuffix(pFile, filepath.Ext(pFile))
		pQuotes := base + qris.ParsedSuffix

		// File to store discarded lines
		pDiscard := base + qris.DiscardSuffix

		pf := qris.ParseFile(pFile)
		qris.WriteDiscards(pf.Discards, pDiscard)
		qris.WriteQuotes(&pf, pQuotes)
	}
	if *validate && allPassed {
		fmt.Println("All files were valid UTF8.")
	}
}
