package main

import (
	"os"

	"github.com/vupdivup/typelines/internal/config"
	"github.com/vupdivup/typelines/internal/domain"
)

func main() {
	if err := config.Init(); err != nil {
		panic(err)
	}
	if err := config.PurgeCache(); err != nil {
		panic(err)
	}

	dirPath := os.Args[1]
	if err := domain.ProcessDirectory(dirPath); err != nil {
		panic(err)
	}
}
