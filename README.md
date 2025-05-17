# Qris

## A Simple Tool for Creating RIS Files

Qris is designed to process files containing annotated quotations into a RIS format suitable for import into [EndNote](https://support.clarivate.com/Endnote/). A text editor or word processor can be used to create and maintain quote files which may later be processed by Qris and easily imported into EndNote.

The tool can accept either `.txt` or `.docx` files as input. The input files may contain any number of source citations, and each source may be associated with any number of quotes. Each quote is processed into a RIS citation record; the RIS records are collected into `.ris` output files.

The current input annotation format is specific to a particular use case, but this may become configurable in the future.

See the [Qris Wiki](https://github.com/paralogismos/qris/wiki) for more detailed information.

## Simple Installation

Grab the `qris.exe` executable from the [Latest Release](https://github.com/paralogismos/qris/releases/latest) page. Drop the executable into your `/Windows/System32/` folder.

More detailed information can be found on the [Installation](https://github.com/paralogismos/qris/wiki/Installation) wiki page.

## Using the Tool

See the [Qris wiki](https://github.com/paralogismos/qris/wiki) for information about [basic usage](https://github.com/paralogismos/qris/wiki#qris-basic-usage), [command line flags](https://github.com/paralogismos/qris/wiki/Qris-Command-Line-Flags), [quote file formatting](https://github.com/paralogismos/qris/wiki/Qris-Input-and-Output-Formats), and more.

## Known Issues

- The older `.doc` file format used by Microsoft Word before 2007 is not currently supported by Qris.
  - To work around this, open `.doc` files in a word processor that supports them and save them again as `.docx` files.

- Be aware that specially formatted elements in a `.docx` file may not be captured by Qris.
  - This is likely to manifest as missing content that is not reported in a _DISCARD file.
  - Currently Qris can extract html links and non-breaking hyphens, but not numbered or bulleted lists.
  - It would be best to avoid such specially formatted elements, but if you encounter such a problem you might raise an issue on the [Issues](https://github.com/paralogismos/qris/issues) page.
