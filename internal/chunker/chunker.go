package chunker

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"os"
	"sync"

	"github.com/sayuyere/storageX/internal/config"
	errorx "github.com/sayuyere/storageX/internal/errors" // unified errors
)

// Chunk represents a file chunk with metadata
const ChunkMetadataSize = 32 + 8 + 8 // checksum + N + Index

type Chunk struct {
	Data     []byte   // variable length
	N        uint64   // 8 bytes
	Err      error    // not serialized
	Name     string   // not serialized in Bytes()
	Checksum [32]byte // 32 bytes (raw SHA256)
	Index    uint64   // 8 bytes
}

// Bytes serializes the chunk into: checksum | N | Index | data
func (c *Chunk) Bytes() []byte {
	buf := make([]byte, 32+8+8+len(c.Data))
	copy(buf[:32], c.Checksum[:])
	binary.BigEndian.PutUint64(buf[32:40], c.N)
	binary.BigEndian.PutUint64(buf[40:48], c.Index)
	copy(buf[48:], c.Data)
	return buf
}

// ChunkFromBytes reconstructs a Chunk, returns nil if data is invalid
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

// NewFileChunker creates a new FileChunker
func NewFileChunker(chunkSize int) *FileChunker {
	return &FileChunker{ChunkSize: chunkSize}
}

// GetChunker returns the singleton instance
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
		panic(errorx.ErrConfigNotLoaded.Error()) // keep panic for fatal config issues
	}
	return GetChunker(cfg.ChunkSize)
}

// ChunkFileStream streams file chunks of the given size
func (fc *FileChunker) ChunkFileStream(file *os.File) (<-chan Chunk, error) {
	ch := make(chan Chunk)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrFileInfoFetchFailed, err)
	}

	fileName := fileInfo.Name()
	go func() {
		defer file.Close()
		defer close(ch)
		metaSize := ChunkMetadataSize
		dataSize := fc.ChunkSize - metaSize
		buf := make([]byte, dataSize)
		index := uint64(0)

		for {
			n, err := file.Read(buf)
			if n > 0 {
				chunkData := make([]byte, n)
				copy(chunkData, buf[:n])
				hash := sha256.Sum256(chunkData)
				name := fileName + "-chunk-" + uintToString(index)

				ch <- Chunk{
					Data:     chunkData,
					N:        uint64(n),
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
				ch <- Chunk{
					Data:  nil,
					N:     0,
					Err:   errorx.Wrap(errorx.ErrChunkReadFailed, err),
					Name:  "",
					Index: index,
				}
				break
			}
		}
	}()
	return ch, nil
}

// ChunkBytes splits an in-memory byte slice into chunks
func (fc *FileChunker) ChunkBytes(data []byte, fileName string) []Chunk {
	metaSize := ChunkMetadataSize
	dataSize := fc.ChunkSize - metaSize
	var chunks []Chunk

	for i, offset := 0, 0; offset < len(data); i, offset = i+1, offset+dataSize {
		end := offset + dataSize
		if end > len(data) {
			end = len(data)
		}
		chunkData := data[offset:end]
		hash := sha256.Sum256(chunkData)
		name := fileName + "-chunk-" + uintToString(uint64(i))

		chunks = append(chunks, Chunk{
			Data:     chunkData,
			N:        uint64(len(chunkData)),
			Name:     name,
			Checksum: hash,
			Index:    uint64(i),
		})
	}
	return chunks
}

// uintToString avoids fmt.Sprintf in hot paths
func uintToString(u uint64) string {
	// simple, allocation-light number to string conversion
	return string(intToBytes(u))
}

func intToBytes(n uint64) []byte {
	if n == 0 {
		return []byte{'0'}
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return b[i:]
}
