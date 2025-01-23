package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/A-PseudoCode-A/grpc_sso/internal/domain/models"
	"github.com/A-PseudoCode-A/grpc_sso/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// Делаем конструктор Storage
func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) error {
	const op = "storage.sqlite.SaveUser"

	// Подготовка запроса
	stmt, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	// Выполнение запроса с контекстом
	_, err = stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		// Преобразуем ошибку к ошибке sqlite3 и проверяем ее на ошибку уже существ. юзера
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sqlite.User"

	// Подготовка запроса
	stmt, err := s.db.Prepare("SELECT * FROM users WHERE email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s, %w", op, err)
	}

	// Выполнение запроса с контекстом
	var resUser models.User

	err = stmt.QueryRowContext(ctx, email).Scan(&resUser.ID, &resUser.Email, &resUser.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return resUser, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"

	// Подготовка запроса
	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("%s, %w", op, err)
	}

	// Выполнение запроса с контекстом
	var isAdmin bool

	err = stmt.QueryRowContext(ctx, userID).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}