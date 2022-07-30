# qris
A Simple Tool for Parsing Quotations

## Simple Installation
Drop the `qris.exe` binary into the `/Windows/System32/` folder.

## Using the Tool
Open a terminal box.

You can type `qris` , `qris -h`, or `qris -help` to see the following help menu:

```
$ qris -help
Usage of qris:
  -b string
        Path to a directory containing files to be parsed, absolute or relative.
  -config
        Show path to configuration file.
  -d string
        Set the current working directory.
  -f string
        Path to a file to be parsed, absolute or relative.
  -v    Validate UTF8 files.
```

### Setting a Working Directory
Whenever the `qris` command is invoked, the `qris` version number and current working directory are displayed.

Type `qris -d` to set a working directory. The argument to the `-d` flag can be either an absolute path or a path relative to the current working directory. If no working directory is set, `qris` uses whatever working directory the system has assigned to your terminal window.

Once you set a working directory, the path to the directory is saved in a configuration file stored at a system-specific location. You can use `qris -config` to see this location displayed.

Place a workspace folder in your working directory. This folder should contain any files to be parsed. Below, `<directory path>` is the path to a workspace folder which is assumed to be under your working directory.

`Qris` will not parse its own output files. This means that you can run `qris` multiple times in the same directory without needing to delete previous results or polluting the directory with extraneous files.

### Parsing One File
To parse one file, type `qris -f <file path>`.

Here `<file path>` can be an absolute path, or a path relative to your working directory, but in either case the path must lead to an actual file to be parsed.

Two output files will be created in the workspace folder. One `PARSED.ris` file will contain the result of parsing the input file, and one `DISCARD.txt` file will contain any lines which were not parsed. Each unparsed line is preceded by a line number indicating where the line may be found in the original unparsed file.

For example, `qris -f workspace/myFileUTF8.txt` would parse the `myFileUTF8.txt` file found in the `workspace` folder found in your working directory, placing the output in the `workspace` folder.

### Batch Parsing Files
To parse a batch of files, type `qris -b <directory path>`.

Here `<directory path>` can be an absolute path, or a path relative to your workspace directory, but in either case the path must lead to a workspace folder which contains files to be parsed.

The `-b` flag also accepts a dot argument, `.`, which indicates that the files in the working directory are to be parsed. Thus, `qris -b batch` tells `qris` to parse the files in the `batch` folder under the working directory, but `qris -b .` tells `qris` to parse the files found directly in the working directory.

All files found in `<directory path>` will be parsed. Two output files, one `PARSED.ris` file and one `DISCARD.txt` file, will be created for each input file and placed in the workspace folder.

For example, `qris -b workspace/batch` would parse all files found in the `batch` subdirectory of the `workspace` folder found in your home directory.

### Validating UTF8 Files
To verify that a file is valid UTF8, use the `-v` flag combined with either `-f` or `-b`. For example, `qris -v -b batch` would verify and parse all files found in the `batch` directory under the working directory. The file or files will be verified before parsing, and the number of the first failing line will be displayed for any failing files. Even failing files will be parsed.
