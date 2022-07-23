# qris
A Simple Tool for Parsing Quotations

## Simple Installation
Drop the `qris.exe` binary into the `/Windows/System32/` folder.

## Using the Tool
Open a terminal box.

You can type `qris` , `qris -h`, or `qris -help` to see a help menu.

Type `qris -home` to see the path to your home directory.

Place a workspace folder in your home directory. This folder should contain any files to be parsed. Below, `<directory path>` is the path to a workspace folder which is assumed to be under your home directory. Note that this path _must not include the name of your home directory_. In other words, `<directory path>` is the path to your workspace folder relative to your home directory. 

### Parsing One File
To parse one file, type `qris <file path>`.

Two output files will be created in the workspace folder. One `.ris` file will contain the result of parsing the input file, and one `DISCARD.txt` file will contain any lines which were not parsed. Each unparsed line is preceded by a line number indicating where the line may be found in the original unparsed file.

For example, `qris workspace/myFileUTF8.txt` would parse the `myFileUTF8.txt` file found in the `workspace` folder found in your home directory, placing the output in the `workspace` folder.

### Batch Parsing Files
To parse a batch of files, type `qris -batch <directory path>`.

All files found in `<directory path>` will be parsed. Two output files, one `.ris` file and one `DISCARD.txt` file, will be created for each input file and placed in the workspace folder.

For example, `qris -batch workspace/batch` would parse all files found in the `batch` subdirectory of the `workspace` folder found in your home directory.

### Validating UTF8 Files
To verify that a file is valid UTF8, use `qris -valid <filepath>`. You can batch-verify files by using `qris -valid <directory path>`. The file or files will be verified without parsing, and the number of the first failing line will be displayed for any failing files.
