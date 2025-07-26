package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/storageX/internal/chunker"
	"github.com/yourusername/storageX/internal/cloud"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	flag.Parse()

	fmt.Println("Starting storageX...")
	fmt.Println("Using config:", *configPath)

	// Load config (placeholder)
	// config := LoadConfig(*configPath)

	// Example: chunk a file and upload to cloud providers
	chunks, err := chunker.ChunkFile("example.txt", 1024*1024) // 1MB chunks
	if err != nil {
		fmt.Println("Chunking error:", err)
		os.Exit(1)
	}

	// Example: upload to AWS S3 (implement your credentials/config)
	aws := cloud.NewS3Storage()
	for i, chunk := range chunks {
		err := aws.UploadChunk(fmt.Sprintf("chunk-%d", i), chunk)
		if err != nil {
			fmt.Println("Upload error:", err)
		}
	}

	// Example: upload to GCP (implement your credentials/config)
	gcp := cloud.NewGCPStorage()
	for i, chunk := range chunks {
		err := gcp.UploadChunk(fmt.Sprintf("chunk-%d", i), chunk)
		if err != nil {
			fmt.Println("Upload error:", err)
		}
	}
}
