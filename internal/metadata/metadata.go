package metadata

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sayuyere/storageX/internal/config"
	errorx "github.com/sayuyere/storageX/internal/errors" // central error constants
)

type ChunkMetadata struct {
	ChunkName string
	Size      int64
	Checksum  string
	Index     int
	Storage   string
	FileName  string
}

type FileMetadata struct {
	FileName  string
	TotalSize int64
}

type MetadataService struct {
	db   *sql.DB
	lock sync.RWMutex
}

func NewMetadataService(dbPath string) (*MetadataService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrMetadataDBOpenFailed, err)
	}
	if err := initSchema(db); err != nil {
		return nil, errorx.Wrap(errorx.ErrMetadataSchemaInitFailed, err)
	}
	return &MetadataService{db: db}, nil
}

func NewMetadataServiceFromConfig() (*MetadataService, error) {
	db, err := sql.Open("sqlite3", config.GetConfig().Meta.DBPath)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrMetadataDBOpenFailed, err)
	}
	if err := initSchema(db); err != nil {
		return nil, errorx.Wrap(errorx.ErrMetadataSchemaInitFailed, err)
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
	if err != nil {
		return errorx.Wrap(errorx.ErrMetadataSchemaInitFailed, err)
	}
	return nil
}

func (m *MetadataService) AddChunk(fileName string, meta ChunkMetadata) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	// Check if chunk already exists
	row := m.db.QueryRow(`SELECT 1 FROM chunks WHERE chunk_name = ?`, meta.ChunkName)
	var exists int
	err := row.Scan(&exists)
	if err == nil {
		return errorx.WrapWithDetails(errorx.ErrChunkAlreadyExists, meta.ChunkName)
	}
	if err != sql.ErrNoRows && err != nil {
		return errorx.Wrap(errorx.ErrDBQueryFailed, err)
	}

	// Insert chunk
	_, err = m.db.Exec(`INSERT OR REPLACE INTO chunks (chunk_name, file_name, size, checksum, idx, storage) VALUES (?, ?, ?, ?, ?, ?)`,
		meta.ChunkName, fileName, meta.Size, meta.Checksum, meta.Index, meta.Storage)
	if err != nil {
		return errorx.Wrap(errorx.ErrChunkInsertFailed, err)
	}

	// Ensure file entry exists
	_, err = m.db.Exec(`INSERT OR IGNORE INTO files (file_name, total_size) VALUES (?, 0)`, fileName)
	if err != nil {
		return errorx.Wrap(errorx.ErrFileInsertFailed, err)
	}

	// Update total size
	_, err = m.db.Exec(`UPDATE files SET total_size = total_size + ? WHERE file_name = ?`, meta.Size, fileName)
	if err != nil {
		return errorx.Wrap(errorx.ErrFileUpdateFailed, err)
	}
	return nil
}

func (m *MetadataService) AddFile(fileName string, fileSize uint64) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, err := m.db.Exec(`INSERT OR IGNORE INTO files (file_name, total_size) VALUES (?, ?)`, fileName, fileSize)
	if err != nil {
		return errorx.Wrap(errorx.ErrFileInsertFailed, err)
	}
	return nil
}

func (m *MetadataService) GetChunk(chunkName string) (ChunkMetadata, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	row := m.db.QueryRow(`SELECT chunk_name, file_name, size, checksum, idx, storage FROM chunks WHERE chunk_name = ?`, chunkName)
	var meta ChunkMetadata
	if err := row.Scan(&meta.ChunkName, &meta.FileName, &meta.Size, &meta.Checksum, &meta.Index, &meta.Storage); err != nil {
		return ChunkMetadata{}, false
	}
	return meta, true
}

func (m *MetadataService) GetFile(fileName string) (FileMetadata, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	row := m.db.QueryRow(`SELECT file_name, total_size FROM files WHERE file_name = ?`, fileName)
	var meta FileMetadata
	if err := row.Scan(&meta.FileName, &meta.TotalSize); err != nil {
		return FileMetadata{}, false
	}
	return meta, true
}

func (m *MetadataService) ListChunks(fileName string) ([]ChunkMetadata, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	rows, err := m.db.Query(`SELECT chunk_name, file_name, size, checksum, idx, storage FROM chunks WHERE file_name = ? ORDER BY idx`, fileName)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrDBQueryFailed, err)
	}
	defer rows.Close()

	var result []ChunkMetadata
	for rows.Next() {
		var meta ChunkMetadata
		if err := rows.Scan(&meta.ChunkName, &meta.FileName, &meta.Size, &meta.Checksum, &meta.Index, &meta.Storage); err != nil {
			return nil, errorx.Wrap(errorx.ErrDBScanFailed, err)
		}
		result = append(result, meta)
	}
	return result, nil
}

func (m *MetadataService) ListFiles() ([]FileMetadata, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	rows, err := m.db.Query(`SELECT file_name, total_size FROM files`)
	if err != nil {
		return nil, errorx.Wrap(errorx.ErrDBQueryFailed, err)
	}
	defer rows.Close()

	var result []FileMetadata
	for rows.Next() {
		var meta FileMetadata
		if err := rows.Scan(&meta.FileName, &meta.TotalSize); err != nil {
			return nil, errorx.Wrap(errorx.ErrDBScanFailed, err)
		}
		result = append(result, meta)
	}
	return result, nil
}

func (m *MetadataService) DeleteFile(fileName string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, err := m.db.Exec(`DELETE FROM chunks WHERE file_name = ?`, fileName)
	if err != nil {
		return errorx.Wrap(errorx.ErrChunkDeleteFailed, err)
	}
	_, err = m.db.Exec(`DELETE FROM files WHERE file_name = ?`, fileName)
	if err != nil {
		return errorx.Wrap(errorx.ErrFileDeleteFailed, err)
	}
	return nil
}

func (m *MetadataService) DeleteChunk(chunkName string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, err := m.db.Exec(`DELETE FROM chunks WHERE chunk_name = ?`, chunkName)
	if err != nil {
		return errorx.Wrap(errorx.ErrChunkDeleteFailed, err)
	}
	return nil
}

func (m *MetadataService) ChunkExists(chunkName string) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	row := m.db.QueryRow(`SELECT 1 FROM chunks WHERE chunk_name = ?`, chunkName)
	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errorx.Wrap(errorx.ErrDBQueryFailed, err)
	}
	return exists == 1, nil
}

func (m *MetadataService) FileExists(fileName string) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	row := m.db.QueryRow(`SELECT 1 FROM files WHERE file_name = ?`, fileName)
	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errorx.Wrap(errorx.ErrDBQueryFailed, err)
	}
	return exists == 1, nil
}
