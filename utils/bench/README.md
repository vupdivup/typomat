# Benchmark

This is a command-line app that executes typeline's directory processing
operation for benchmarking.
It takes a directory path as a positional argument and exists when said
directory has been ingested.
The database is purged on each run.

## Usage

To get started, first build the workload program.

```bash
go build -o bench.exe ./utils/bench
```

Then, run the program with a benchmarking tool like [hyperfine](https://github.com/sharkdp/hyperfine), providing a directory path as the only argument. Since processing involves a lot of I/O, make sure to request a warmup run as well.

```bash
hyperfine -w 1 'bench.exe [DIRECTORY]'
```

The contents of the following repos, ordered from largest to smallest, are good candidates for benchmarking:

- [microsoft/vscode](https://github.com/microsoft/vscode)
- [logseq/logseq](https://github.com/logseq/logseq)
- [tweenjs/tween.js](https://github.com/tweenjs/tween.js.git)
