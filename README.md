# qris
A Simple Tool for Parsing Quotations

## Simple Installation
Drop the `qris.exe` binary into the `/Windows/System32/` folder.

## Using the Tool
Open a terminal box.

You can type `qris` to see a help menu.

Type `qris -h` to see the path to your home directory.

Place a workspace folder in your home directory. This folder should contain any files to be parsed. Below, `workspace` is the path to a workspace folder which is assumed to be in your home directory. Note that this path _must not include the name of your home directory_. In other words, `workspace` is the path to your workspace folder relative to your home directory.

### Parsing One File
To parse one file, type, e.g. `qris workspace/myFileUTF8.txt`.

Two output files will be created in the `workspace` folder. One `.ris` file will contain the result of parsing the input file, and one `DISCARD.txt` file will contain any lines which were not parsed. Each unparsed line is preceded by a line number indicating where the line may be found in the original unparsed file.

### Batch Parsing Files
To parse a batch of files, type, e.g., `qris -b workspace`. All files found in `workspace` will be parsed. Two output files, one `.ris` file and one `DISCARD.txt` file, will be created for each input file and placed in the `workspace` folder.
