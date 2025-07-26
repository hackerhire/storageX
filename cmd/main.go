package main

import (
	"flag"
	"fmt"

	"github.com/sayuyere/storageX/internal/log"
)

func main() {
	log.InitLogger(true) // log only to stdout, debug enabled
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	flag.Parse()

	fmt.Println("Starting storageX...")
	fmt.Println("Using config:", *configPath)

	// Load config (placeholder)
	// config := LoadConfig(*configPath)

	// Example: chunk a file and upload to Google Drive

}
