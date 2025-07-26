package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sayuyere/storageX/internal/chunker"
	"github.com/sayuyere/storageX/internal/cloud"
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
	chunks, err := chunker.ChunkFile("example.txt", 1024*1024) // 1MB chunks
	if err != nil {
		fmt.Println("Chunking error:", err)
		os.Exit(1)
	}

	drive := cloud.NewDriveStorage()
	for i, chunk := range chunks {
		err := drive.UploadChunk(fmt.Sprintf("chunk-%d", i), chunk)
		if err != nil {
			fmt.Println("Upload error:", err)
		}
	}
}
