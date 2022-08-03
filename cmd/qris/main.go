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
	"os"

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

	// Set a new working directory if needed.
	qris.SetWorkDir(*dir, configPath)

	// Get current working directory.
	workDir := qris.GetWorkDir(configPath)

	// Always show current qris version and current working directory
	fmt.Println("qris version", qris.Version)
	fmt.Println("Working in directory", workDir)

	// `dataList` is a list of files to be processed.
	//	var dataList []string

	// `workPath` is the absolute path to files to be processed.
	// `batchPath` may include directory structure relative to the
	// working directory, and this additional directory structure is
	// included in `workPath`.
	dataList, workPath := qris.GetWorkPath(workDir, *batchPath, *filePath)

	// Parse all files and write results to output.
	allPassed := qris.WriteResults(workPath, dataList, *validate)

	if *validate && allPassed {
		fmt.Println("All files were valid UTF8.")
	}
}
