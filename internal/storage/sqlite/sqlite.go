package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"url-shorter/internal/storage"

	"github.com/mattn/go-sqlite3"
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

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const fn = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error);
		ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil

}

func (s *Storage) GetURL(alias string) (string, error) {
	const fn = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}

	if err != nil {
		return "", fmt.Errorf("%s: execute statement: %w", fn, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(id int) error {
	const fn = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if affected == 0 {
		return fmt.Errorf("%s: %w", fn, storage.ErrURLNotFound)
	}

	return nil
}

func (s *Storage) IsAliasExists(alias string) (bool, error) {
	const fn = "torage.sqlite.IsAliasExists"
	
	stmt, err := s.db.Prepare("SELECT COUNT(*) FROM url WHERE alias = ?")
	if err != nil {
		return false, fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	var count int 
	err = stmt.QueryRow(alias).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, storage.ErrURLNotFound
	}

	if err != nil {
		return false, fmt.Errorf("%s: execute statement: %w", fn, err)
	}
	
	return count > 0, nil
}