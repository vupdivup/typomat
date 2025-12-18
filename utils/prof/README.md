# Profiling

This is a command-line app that executes typeline's directory processing
operation for CPU profiling.
It takes a directory path and an output filename as positional arguments and
generates a CPU profile.
The database is purged on each run.

## Usage

To get started, first build the profiling program.

```bash
go build -o prof.exe ./utils/prof
```

Then, run the program with a directory path and output filename as arguments.

```bash
./prof.exe [DIRECTORY] [OUTPUT_FILE]
```

Analyze the resulting profile with Go's pprof tool:

```bash
go tool pprof -http=:8080 [OUTPUT_FILE]
```
