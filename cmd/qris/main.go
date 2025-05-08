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
	"path/filepath"
	"regexp"

	"qris"
)

// Returns a string representing the name of the executable command.
func command(args_0 string) string {
	cmd := filepath.Base(args_0)
	ext := filepath.Ext(args_0)
	commandExt := regexp.MustCompile(ext + `$`)
	cmd = commandExt.ReplaceAllLiteralString(cmd, "")
	return cmd
}

func main() {

	// Parse command line flags.
	batchPath := flag.String("b", "",
		"Path to a directory containing files to be parsed, absolute or relative.")
	conf := flag.String("config", "",
		"p, path: Show path to configuration file.\nr, rm, remove: Remove configuration file.")
	dir := flag.String("d", "",
		"Set the current working directory.")
	encoding := flag.String("enc", "ext",
		"Output encoding.\nOne of 'ascii', 'ext', 'utf8', or 'utf16'")
	filePath := flag.String("f", "",
		"Path to a file to be parsed, absolute or relative.")
	noDateStamp := flag.Bool("nods", false, "Omit AD datestamp field.")
	volume := flag.Bool("v", false, "Include VL volume field.")

	// Custom usage message.
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n",
			command(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validate -f and -b flag usage.
	if *filePath != "" && *batchPath != "" {
		fmt.Fprintln(os.Stderr, "-f and -b flags may not be used together")
		flag.Usage()
		os.Exit(1)
	}

	// Set encoding.
	var enc qris.Encoding
	switch *encoding {
	case "ascii":
		enc = qris.Ascii
	case "ext":
		enc = qris.ExtendedAscii
	case "utf8":
		enc = qris.Utf8
	case "utf16":
		enc = qris.Utf16
	default:
		enc = qris.Ascii
	}

	// Configure the system.
	configPath := qris.GetConfigPath()
	if *conf == "p" || *conf == "path" {
		fmt.Println("Configuration file at", configPath)
	} else if *conf == "r" || *conf == "rm" || *conf == "remove" {
		err := os.Remove(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Removed %s\n", configPath)
	} else if *conf != "" {
		fmt.Fprintf(os.Stderr, "-config: unrecognized argument '%s'\n", *conf)
		flag.Usage()
		os.Exit(1)
	}

	// Set a new working directory if needed.
	qris.SetWorkDir(*dir, configPath)

	// Get current working directory.
	workDir := qris.GetWorkDir(configPath)

	// Always show current qris version and current working directory
	fmt.Println("qris version", qris.Version)
	fmt.Println("Working in directory", workDir)

	// `workPath` is the absolute path to files to be processed.
	// `batchPath` may include directory structure relative to the
	// working directory, and this additional directory structure is
	// included in `workPath`.
	dataList, workPath := qris.GetWorkPath(workDir, *batchPath, *filePath)

	// Parse all files and write results to output.
	qris.WriteResults(workPath, dataList, *volume, *noDateStamp, enc)
}
