# Simple Real-Time Translator Tool (srtt)

Simple Real-Time Translator Tool (srtt) is a command-line interface (CLI) application that provides real-time
translation using various translation engines like DeepL and Baidu. It's designed to be easy to use for translating text
files, specifically subtitles in SRT format, from one language to another.

## Features

- Support for multiple translation engines (DeepL, Baidu)
- Configuration via CLI arguments or environment variables
- Customizable translation context length and offset
- Debug mode for troubleshooting
- Automatic retry mechanism for handling API request failures

## Installation

To install srtt, you need to have Go installed on your machine. Follow these steps to install srtt:

```bash
git clone https://github.com/AnTengye/srtt.git
cd srtt
go build -o srtt
```

## Usage

To use srtt, you can run it directly from the command line. Here's how you can perform a translation:

```bash
./srtt translate --source ja --target zh --input yourfile.srt --output translatedfile.srt
```

### Command-Line Arguments

- --source, -s: Source language (default "ja")
- --target, -t: Target language (default "zh")
- --input, -i: Input file path (default "in.srt")
- --output, -o: Output file path
- --engine, -e: Translation engine ("deeplx", "baidu"; default "deeplx")
  Additional flags for customization like --debug for enabling debug mode, --retry for setting retry attempts, etc.

## Contributing

Contributions to srtt are welcome! Please fork the repository and submit pull requests with any improvements or bug
fixes.

## License

srtt is open-sourced under the MIT license. See the LICENSE file for more details.