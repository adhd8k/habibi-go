package main

import (
	"embed"
	"log"
	"os"

	"habibi-go/cmd"
)

//go:embed web/dist/*
var webAssets embed.FS

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}