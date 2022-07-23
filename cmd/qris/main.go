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
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"qris"
	"strings"
)

func main() {
	// Get user's home directory.
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Parse command line flags
	home := flag.Bool("home", false, "Show user's home directory")
	batch := flag.Bool("batch", false, "Batch process files")
	valid := flag.Bool("valid", false, "Validate UTF8 files")
	flag.Parse()

	// Collect path information from command line.
	var inputPath string
	var absInputPath string
	var dataPath string
	isPath := false
	if len(flag.Args()) > 0 {
		isPath = true
		inputPath = flag.Args()[0]
		absInputPath = filepath.Join(usr.HomeDir, inputPath)
		// Initialize `dataPath` assuming single file (-batch is false)
		dataPath = filepath.Dir(absInputPath)
	}

	switch {
	case *home:
		// User requests home directory display.
		fmt.Println(usr.HomeDir) // display home directory
		os.Exit(0)
	case *batch:
		// User wants to batch-process files.
		dataPath = absInputPath // path to folder containing files
	case *valid:
		fallthrough
	default:
		// No flags are acceptable only when a filepath is provided.
		if !isPath {
			flag.Usage()
			os.Exit(1)
		}
	}

	// list of files to be processed
	var dataList []string

	if *batch {
		files, err := os.ReadDir(absInputPath)
		if err != nil {
			log.Fatal(err)
		}
		// create list of files in working directory
		for _, file := range files {
			dataList = append(dataList, file.Name())
		}
	} else {
		// create list containing only one file from command-line
		dataList = append(dataList, filepath.Base(absInputPath))
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
			vFile := filepath.Join(dataPath, file)
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
			pFile := filepath.Join(dataPath, file)
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
