package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"

	"url-shorter/internal/lib/hash_password"
	"url-shorter/internal/storage"
)

func (s *Storage) SaveUser(username, email, password string) (int64, error) {
	const fn = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO user(username, email, password, created_at) VALUES(?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	hashPassword, err := hash_password.GeneratePassword(password)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	createdAt := time.Now().UTC()
	res, err := stmt.Exec(username, email, hashPassword, createdAt)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				if strings.Contains(err.Error(), "username") {
					return 0, fmt.Errorf("%s: username уже существует: %w", fn, storage.ErrUsernamelExists)
				}
				if strings.Contains(err.Error(), "email") {
					return 0, fmt.Errorf("%s: email уже существует: %w", fn, storage.ErrEmailExists)
				}
			}
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) ValidateUser(username, password string) (bool, error) {
	const fn = "storage.sqlite.ValidateUser"
	var hashPassword string

	stmt, err := s.db.Prepare("SELECT password FROM user WHERE username = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", fn, err)
	}
	err = stmt.QueryRow(username).Scan(&hashPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	// Сравниваем переданный пароль с хэшированным значением
	err = bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	if err != nil {
		return false, storage.ErrInvalidPassword // Пароль неверный
	}

	return true, nil // Валидация успешна
}
