package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"

	"url-shorter/internal/storage"
)

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const fn = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
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

	tx, err := s.db.Begin()
	if err != nil {
		return "", fmt.Errorf("%s: failed to start transaction: %w", fn, err)
	}

	stmt, err := tx.Prepare(`
        UPDATE url
        SET clicks = clicks - 1
        WHERE alias = ? AND clicks > 0
        RETURNING url;
    `)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}
	defer stmt.Close()

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: query failed: %w", fn, err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("%s: failed to commit transaction: %w", fn, err)
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
	const fn = "storage.sqlite.IsAliasExists"

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
