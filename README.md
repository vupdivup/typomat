# typomat: turn your code into muscle memory

typomat is a command-line typing practice tool that creates exercises from the contents of your repository — perfect for a quick warmup before work.

<figure>
    <img src="docs/demo.gif"
         alt="Typing session demo">
    <figcaption><center><em>Practicing with words from the typomat project's source code</em></center></figcaption>
</figure>

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
scoop install vupdivup/typomat
```

### Linux/MacOS

Install via [Homebrew](https://brew.sh/):

```bash
brew tap vupdivup/tap
brew install typomat
```

### Go

Alternatively, if Go is available on your system:

```bash
go install github.com/vupdivup/typomat/cmd/typomat@latest
```

Note that this method may require manually adding the [GOBIN](https://go.dev/wiki/GOPATH#gopath-variable) directory to PATH.
