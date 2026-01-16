# typomat: turn your code into muscle memory

typomat is a command-line typing practice tool that creates exercises from the contents of your repository—perfect for a quick warmup before work.

![typomat demo](docs/demo.gif)

## Features

- ⌨️ Typing practice focused on code vocabulary
- 📂 Works with any local folder containing text files
- 🖥️ Sleek text-based UI that fits comfortably in your terminal
- 📊 WPM and accuracy metrics to help track your progress
- 🙈 .gitignore-aware; no sensitive data is ingested
- 💾 Caching to load your favorite codebases in no time

## How it works

When invoked, typomat runs through a directory's source code, extracting words from variable declarations, string literals and function signatures. These words are then used to build short, randomized typing prompts relevant to your codebase.

## Installation

### Windows

Install via [Scoop](https://scoop.sh/):

```bash
scoop bucket add vupdivup https://github.com/vupdivup/scoop-bucket
scoop install typomat
```

### Linux/MacOS

Install via [Homebrew](https://brew.sh/):

```bash
brew tap vupdivup/tap
brew install typomat
```

### Pre-built binaries

Pre-built binaries are archived under the [latest release](https://github.com/vupdivup/typomat/releases/latest).

### Go

Use the Go toolchain to build from source:

```bash
go install github.com/vupdivup/typomat/cmd/typomat@latest
```

## Usage

Start a typing session by passing the path to the directory you'd like to practice on:

```bash
typomat path/to/dir
```

For large directories, startup times can be greatly reduced by reusing data across sessions. Pass the `--cache` flag to store results for subsequent runs:

```bash
typomat --cache path/to/dir
```
 