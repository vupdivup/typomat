package main

import (
	"os"
	"runtime/pprof"

	"github.com/vupdivup/typomat/internal/config"
	"github.com/vupdivup/typomat/internal/domain"
)

func main() {
	if err := config.Init(); err != nil {
		panic(err)
	}
	if err := config.PurgeCache(); err != nil {
		panic(err)
	}

	filename := os.Args[2]
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	dirPath := os.Args[1]
	if err := domain.ProcessDirectory(dirPath); err != nil {
		panic(err)
	}
	pprof.StopCPUProfile()
}
