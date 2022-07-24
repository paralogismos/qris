# qris
A Simple Tool for Parsing Quotations

## Simple Installation
Drop the `qris.exe` binary into the `/Windows/System32/` folder.

## Using the Tool
Open a terminal box.

You can type `qris` , `qris -h`, or `qris -help` to see the following help menu:

```
Usage of qris:
  -b string
        Path to a directory containing files to be parsed, relative to working directory.
  -config
        Show path to configuration file.
  -d string
        Set the current working directory.
  -f string
        Path to a file to be parsed, relative to working directory.
  -v    Validate UTF8 files.
```

Whenever the `qris` command is invoked, the current working directory is displayed.

Type `qris -d` to set a working directory. The argument to the `-d` flag can be either an absolute path or a path relative to the current working directory. If no working directory is set, `qris` uses whatever working directory the system has assigned to your terminal window.

Once you set a working directory, the path to the directory is saved in a configuration file stored at a system-specific location. You can use `qris -config` to see this location displayed.

Place a workspace folder in your working directory. This folder should contain any files to be parsed. Below, `<directory path>` is the path to a workspace folder which is assumed to be under your working directory.

### Parsing One File
To parse one file, type `qris -f <file path>`.

Here `<file path>` can be an absolute path, or a path relative to your working directory, but in either case the path must lead to an actual file to be parsed.

Two output files will be created in the workspace folder. One `.ris` file will contain the result of parsing the input file, and one `DISCARD.txt` file will contain any lines which were not parsed. Each unparsed line is preceded by a line number indicating where the line may be found in the original unparsed file.

For example, `qris -f workspace/myFileUTF8.txt` would parse the `myFileUTF8.txt` file found in the `workspace` folder found in your working directory, placing the output in the `workspace` folder.

### Batch Parsing Files
To parse a batch of files, type `qris -b <directory path>`.

Here `<directory path>` can be an absolute path, or a path relative to your workspace directory, but in either case the path must lead to a workspace folder which contains files to be parsed.

All files found in `<directory path>` will be parsed. Two output files, one `.ris` file and one `DISCARD.txt` file, will be created for each input file and placed in the workspace folder.

For example, `qris -b workspace/batch` would parse all files found in the `batch` subdirectory of the `workspace` folder found in your home directory.

### Validating UTF8 Files
To verify that a file is valid UTF8, use the `-v` flag combined with either `-f` or `-b`. For example, `qris -v -b batch` would verify and parse all files found in the `batch` directory under the working directory. The file or files will be verified before parsing, and the number of the first failing line will be displayed for any failing files. Even failing files will be parsed.
