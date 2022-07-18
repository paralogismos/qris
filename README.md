# qris
A Simple Tool for Parsing Quotations

## Simple Installation
Drop the `qris.exe` binary into the `/Windows/System32/` folder.

## Using the Tool
Open a terminal box.

You can type `qris` to see a help menu.

Type `qris -h` to see the path to your home directory.

Place a workspace folder in your home directory. This folder should contain any files to be parsed. Below, `workspace` is the name of the workspace folder which is assumed to be in your home directory.

### Parsing One File
To parse one file, type, e.g. `qris workspace/myFileUTF8.txt`.

Two output files will be created in the `workspace` folder. One `.ris` file will contain the result of parsing the input file, and one `DISCARD.txt` file will contain any lines which were not parsed. Each unparsed line is preceded by a line number indicating where the line may be found in the original unparsed file.

### Batch Parsing Files
If you want to parse multiple files at once, create a text file containing the name of each file to be parsed. Each line should contain only one name. Only the name of the file should be used; `qris` assumes that the file list and the files to be parsed exist in the same folder.

To parse a batch of files, type, e.g., `qris -b workspace/fileList.txt`.

Two output files, one `.ris` file and one `DISCARD.txt` file, will be created for each input file and placed in the `workspace` folder.
