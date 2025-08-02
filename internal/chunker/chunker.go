package chunker

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sayuyere/storageX/internal/config"
)

// Chunk represents a file chunk with metadata
// Data: chunk bytes
// N: number of bytes read
// Err: error if any
// Name: unique chunk name (e.g., file-chunk-0)
// Checksum: fixed-length SHA256 (32 bytes, hex-encoded)
// Index: chunk index in file

const ChunkMetadataSize = 32 + 8 + 8 // checksum + N + Index

type Chunk struct {
	Data     []byte   // variable length
	N        uint64   // 8 bytes
	Err      error    // not serialized
	Name     string   // not serialized in Bytes()
	Checksum [32]byte // 32 bytes (raw SHA256)
	Index    uint64   // 8 bytes
}

// Bytes returns the chunk as: [32]byte checksum | uint64 N | uint64 Index | chunk data
func (c *Chunk) Bytes() []byte {
	buf := make([]byte, 32+8+8+len(c.Data))
	copy(buf[:32], c.Checksum[:])
	binary.BigEndian.PutUint64(buf[32:40], c.N)
	binary.BigEndian.PutUint64(buf[40:48], c.Index)
	copy(buf[48:], c.Data)
	return buf
}

// ChunkFromBytes reconstructs a Chunk from a byte slice (inverse of Bytes)
func ChunkFromBytes(b []byte) *Chunk {
	if len(b) < 48 {
		return nil
	}
	var c Chunk
	copy(c.Checksum[:], b[:32])
	c.N = binary.BigEndian.Uint64(b[32:40])
	c.Index = binary.BigEndian.Uint64(b[40:48])
	c.Data = make([]byte, len(b)-48)
	copy(c.Data, b[48:])
	return &c
}

type FileChunker struct {
	ChunkSize int
}

var (
	singleton     *FileChunker
	singletonOnce sync.Once
)

// NewFileChunker creates a new FileChunker with the given chunk size
func NewFileChunker(chunkSize int) *FileChunker {
	return &FileChunker{ChunkSize: chunkSize}
}

// GetChunker returns the singleton FileChunker with the given chunk size (set only on first call)
func GetChunker(chunkSize int) *FileChunker {
	singletonOnce.Do(func() {

		singleton = &FileChunker{ChunkSize: chunkSize}
	})
	return singleton
}

// GetChunkerFromConfig returns the singleton FileChunker using chunk size from app config
func GetChunkerFromConfig() *FileChunker {
	cfg := config.GetConfig()
	if cfg == nil {
		panic("App config not loaded. Call config.LoadConfig first.")
	}
	return GetChunker(cfg.ChunkSize)
}

// ChunkFileStream streams file chunks of the given size (in bytes) via a channel, with metadata
func (fc *FileChunker) ChunkFileStream(file *os.File) (<-chan Chunk, error) {
	ch := make(chan Chunk)
	fileName := file.Name() // get the file name for chunk naming
	go func() {
		defer file.Close()
		defer close(ch)
		metaSize := 32 + 8 + 8 // checksum + N + Index
		dataSize := fc.ChunkSize - metaSize
		buf := make([]byte, dataSize)
		index := uint64(0)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				hash := sha256.Sum256(chunk)
				name := fmt.Sprintf("%s-chunk-%d", fileName, index)
				ch <- Chunk{
					Data:     chunk,
					N:        uint64(n),
					Err:      nil,
					Name:     name,
					Checksum: hash,
					Index:    index,
				}
				index++
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				ch <- Chunk{Data: nil, N: 0, Err: err, Name: "", Index: index}
				break
			}
		}
	}()
	return ch, nil
}

// ChunkBytes splits a byte slice into chunks and returns a slice of Chunk
// This is useful for testing or when you have data in memory
func (fc *FileChunker) ChunkBytes(data []byte, fileName string) []Chunk {
	metaSize := 32 + 8 + 8 // checksum + N + Index
	dataSize := fc.ChunkSize - metaSize
	var chunks []Chunk
	for i, offset := 0, 0; offset < len(data); i, offset = i+1, offset+dataSize {
		end := offset + dataSize
		if end > len(data) {
			end = len(data)
		}
		chunkData := data[offset:end]
		hash := sha256.Sum256(chunkData)
		name := fmt.Sprintf("%s-chunk-%d", fileName, i)
		chunks = append(chunks, Chunk{
			Data:     chunkData,
			N:        uint64(len(chunkData)),
			Err:      nil,
			Name:     name,
			Checksum: hash,
			Index:    uint64(i),
		})
	}
	return chunks
}
