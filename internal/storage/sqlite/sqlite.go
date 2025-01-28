package sqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	const fn = "storage.sqlite.NewStorage"

	// Определяем директорию корня проекта
	rootDir, err := filepath.Abs(filepath.Join("..", "..")) // Поднимаемся из cmd/url-shorter к корню
	if err != nil {
		return nil, fmt.Errorf("%s: failed to resolve root directory: %w", fn, err)
	}
	fmt.Println("Root directory:", rootDir)

	// Создаём абсолютный путь для базы данных
	absPath := filepath.Join(rootDir, dbPath)
	fmt.Println("Absolute database path:", absPath)

	db, err := sql.Open("sqlite3", absPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &Storage{db: db}, nil
}

