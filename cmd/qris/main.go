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
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"qris"
)

func main() {
	// Get user's home directory.
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Process command-line arguments

	// Develop multiple file entry option?

	// Help facility
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println(os.Args[0], "               -- This help screen")
		fmt.Println(os.Args[0], "-h             -- display home directory path")
		fmt.Println(os.Args[0], "[filename]     -- process a single file")
		fmt.Println(os.Args[0], "-b [filename]  -- process a list of files")
		os.Exit(0)
	}

	// Collect user path entry.
	inputPath := ""
	if os.Args[1] == "-h" {
		fmt.Println(usr.HomeDir) // display home directory
		os.Exit(0)
	} else if os.Args[1] == "-b" {
		inputPath = os.Args[2] // path to file containing list of files
	} else {
		inputPath = os.Args[1] // path to single file
	}

	// Obtain list of files to work from
	// In a simple command-line interface, `testInput` should be a parameter
	// taken from the command-line.
	//	testInput := "/rees_quotes/qris_test.txt"

	// path to file or list to be processed
	absInputPath := filepath.Join(usr.HomeDir, inputPath)

	// directory of files to be processed
	dataPath := filepath.Dir(absInputPath)

	// list of files to be processed
	var dataList []string

	if os.Args[1] == "-b" {
		dataList = qris.GetFileList(absInputPath) // parsed from input file
	} else {
		dataList = append(dataList, filepath.Base(absInputPath)) // parsed from command-line
	}

	// Parse all files
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
		/*
			fmt.Println("\n*************************")
			fmt.Println("Quotes file:", pQuotes)
			fmt.Println("Discard file:", pDiscard)
			fmt.Println("File:", pf.Filename)
			fmt.Println("Source:", pf.Title)
			fmt.Println("Citation:", pf.Citation.Body)
			fmt.Println("    Name:", pf.Citation.Name)
			fmt.Println("    Year:", pf.Citation.Year)
			fmt.Println("-------------------------")
			fmt.Println("-        QUOTES         -")
			fmt.Println("-------------------------")
			for _, q := range pf.Quotes {
				fmt.Println(q.Body)
				fmt.Println("    Page:", q.Page)
				fmt.Println("    Supp:", q.Supp, "\n")
			}
		*/
		//		fmt.Println("-------------------------")
		//		fmt.Println("-       DISCARDS        -")
		//		fmt.Println("-------------------------")
		//		for _, d := range pf.Discards {
		//			fmt.Println("<", d.LineNo, ">")
		//			fmt.Println(d.Body, "\n")
		//		}
		//		fmt.Println("*************************\n")
	}
}
