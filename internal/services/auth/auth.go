package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/A-PseudoCode-A/grpc_sso/internal/domain/models"
	"github.com/A-PseudoCode-A/grpc_sso/internal/lib/jwt"
	"github.com/A-PseudoCode-A/grpc_sso/internal/lib/logger/sl"
	"github.com/A-PseudoCode-A/grpc_sso/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUSer(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvdlidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

// New returns a new instance of the Auth service
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Login позволяет залогиняться нашим юзерам
func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))
	log.Info("attemping to login user")

	// Если ошибка соответствует ошибки из storage, то вернем ее, иначе просто обработаем
	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем на правильность пароля
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged is successfully")

	// Для каждого приложения будет свой jwt ключ (точнее, специальная кодовая фаза)
	// Приватный ключ будет хранится и на стороне клиентоского сервиса, и на стороне SSO

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// RegisterNewUser позволяет зарегестрироваться нашим полльзователям
func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(slog.String("op", op), slog.String("email", email))
	log.Info("registering new user")

	//Нужно захэшировать пароль, а также посолить его (добавить в начало несколько нужных строчек)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Сохраняем в БД
	id, err := a.userSaver.SaveUSer(ctx, email, passHash)

	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

// Login позволяет зарегестрироваться нашим полльзователям
func (a *Auth) IsAdmin(ctx context.Context, userID int) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(slog.String("op", op), slog.Int64("userID", int64(userID)))
	log.Info("checking the user for the admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvdlidAppID)
		}
		log.Warn("failed to check on admin", sl.Err(err))

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("the user is an admin", slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}
