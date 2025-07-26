package metadata

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type ChunkMetadata struct {
	ChunkName string
	Size      int64
	Checksum  string
	Index     int
	Storage   string // e.g., provider name or location
	FileName  string // for easier queries
}

type FileMetadata struct {
	FileName  string
	TotalSize int64
}

type MetadataService struct {
	db *sql.DB
}

func NewMetadataService(dbPath string) (*MetadataService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := initSchema(db); err != nil {
		return nil, err
	}
	return &MetadataService{db: db}, nil
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS chunks (
		chunk_name TEXT PRIMARY KEY,
		file_name TEXT,
		size INTEGER,
		checksum TEXT,
		idx INTEGER,
		storage TEXT
	);
	CREATE TABLE IF NOT EXISTS files (
		file_name TEXT PRIMARY KEY,
		total_size INTEGER
	);
	`)
	return err
}

func (m *MetadataService) AddChunk(fileName string, meta ChunkMetadata) error {
	// Check if chunk already exists
	row := m.db.QueryRow(`SELECT 1 FROM chunks WHERE chunk_name = ?`, meta.ChunkName)
	var exists int
	err := row.Scan(&exists)
	if err == nil {
		return fmt.Errorf("chunk %s already exists", meta.ChunkName)
	}
	// Check if file already exists
	row = m.db.QueryRow(`SELECT 1 FROM files WHERE file_name = ?`, fileName)
	err = row.Scan(&exists)
	if err == nil && exists == 1 {
		return fmt.Errorf("file %s already exists", fileName)
	}
	_, err = m.db.Exec(`INSERT OR REPLACE INTO chunks (chunk_name, file_name, size, checksum, idx, storage) VALUES (?, ?, ?, ?, ?, ?)`,
		meta.ChunkName, fileName, meta.Size, meta.Checksum, meta.Index, meta.Storage)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(`INSERT OR IGNORE INTO files (file_name, total_size) VALUES (?, 0)`, fileName)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(`UPDATE files SET total_size = total_size + ? WHERE file_name = ?`, meta.Size, fileName)
	return err
}

func (m *MetadataService) GetChunk(chunkName string) (ChunkMetadata, bool) {
	row := m.db.QueryRow(`SELECT chunk_name, file_name, size, checksum, idx, storage FROM chunks WHERE chunk_name = ?`, chunkName)
	var meta ChunkMetadata
	if err := row.Scan(&meta.ChunkName, &meta.FileName, &meta.Size, &meta.Checksum, &meta.Index, &meta.Storage); err != nil {
		return ChunkMetadata{}, false
	}
	return meta, true
}

func (m *MetadataService) GetFile(fileName string) (FileMetadata, bool) {
	row := m.db.QueryRow(`SELECT file_name, total_size FROM files WHERE file_name = ?`, fileName)
	var meta FileMetadata
	if err := row.Scan(&meta.FileName, &meta.TotalSize); err != nil {
		return FileMetadata{}, false
	}
	return meta, true
}

func (m *MetadataService) ListChunks(fileName string) ([]ChunkMetadata, error) {
	rows, err := m.db.Query(`SELECT chunk_name, file_name, size, checksum, idx, storage FROM chunks WHERE file_name = ? ORDER BY idx`, fileName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ChunkMetadata
	for rows.Next() {
		var meta ChunkMetadata
		if err := rows.Scan(&meta.ChunkName, &meta.FileName, &meta.Size, &meta.Checksum, &meta.Index, &meta.Storage); err != nil {
			return nil, err
		}
		result = append(result, meta)
	}
	return result, nil
}

func (m *MetadataService) ListFiles() ([]FileMetadata, error) {
	rows, err := m.db.Query(`SELECT file_name, total_size FROM files`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []FileMetadata
	for rows.Next() {
		var meta FileMetadata
		if err := rows.Scan(&meta.FileName, &meta.TotalSize); err != nil {
			return nil, err
		}
		result = append(result, meta)
	}
	return result, nil
}

// DeleteFile deletes a file and all its linked chunks from the database
func (m *MetadataService) DeleteFile(fileName string) error {
	// Delete all chunks for this file
	_, err := m.db.Exec(`DELETE FROM chunks WHERE file_name = ?`, fileName)
	if err != nil {
		return err
	}
	// Delete the file entry itself
	_, err = m.db.Exec(`DELETE FROM files WHERE file_name = ?`, fileName)
	return err
}
