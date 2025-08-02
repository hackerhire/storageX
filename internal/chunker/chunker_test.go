package chunker

import (
	"bytes"
	"crypto/sha256"
	"io/ioutil"
	"os"
	"testing"
)

func TestChunk_BytesAndFromBytes(t *testing.T) {
	data := []byte("hello world")
	hash := sha256.Sum256(data)
	chunk := &Chunk{
		Data:     data,
		N:        uint64(len(data)),
		Name:     "file-chunk-0",
		Checksum: hash,
		Index:    0,
	}
	b := chunk.Bytes()
	chunk2 := ChunkFromBytes(b)
	if chunk2 == nil {
		t.Fatal("ChunkFromBytes returned nil")
	}
	if !bytes.Equal(chunk.Data, chunk2.Data) {
		t.Errorf("data mismatch: got %v, want %v", chunk2.Data, chunk.Data)
	}
	if chunk.N != chunk2.N {
		t.Errorf("N mismatch: got %d, want %d", chunk2.N, chunk.N)
	}
	if chunk.Index != chunk2.Index {
		t.Errorf("Index mismatch: got %d, want %d", chunk2.Index, chunk.Index)
	}
	if chunk.Checksum != chunk2.Checksum {
		t.Errorf("Checksum mismatch: got %x, want %x", chunk2.Checksum, chunk.Checksum)
	}
}

func TestFileChunker_ChunkFileStream(t *testing.T) {
	f, err := ioutil.TempFile("", "chunker-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	data := []byte("abcdefghijklmnopqrstuvwxyz")
	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek: %v", err)
	}
	chunker := NewFileChunker(48 + 5) // 5 bytes data per chunk
	ch, err := chunker.ChunkFileStream(f)
	if err != nil {
		t.Fatalf("ChunkFileStream error: %v", err)
	}
	var count int
	for chunk := range ch {
		if chunk.Err != nil {
			t.Errorf("chunk error: %v", chunk.Err)
		}
		if len(chunk.Data) == 0 {
			t.Error("chunk data empty")
		}
		count++
	}
	if count == 0 {
		t.Error("no chunks produced")
	}
}

func TestGetChunker_Singleton(t *testing.T) {
	c1 := GetChunker(100)
	c2 := GetChunker(200)
	if c1 != c2 {
		t.Error("GetChunker did not return singleton instance")
	}
	if c1.ChunkSize != 100 {
		t.Errorf("expected chunk size 100, got %d", c1.ChunkSize)
	}
}
