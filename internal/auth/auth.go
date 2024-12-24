package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists        = errors.New("user with such login already registered")
	ErrIncorrectUserCredentials = errors.New("user credentials isn`t valid")
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Auther interface {
	RegistrateUser(ctx context.Context, ar *AuthRequest) (string, error)
	LoginUser(ctx context.Context, ar *AuthRequest) (string, error)
	AuthMiddleWare(h http.HandlerFunc) http.HandlerFunc
}

type Auth struct {
	key []byte
	db  *sql.DB
}

func NewAuth(key []byte, db *sql.DB) *Auth {

	return &Auth{
		key: key,
		db:  db,
	}
}

func (a *Auth) RegistrateUser(ctx context.Context, ar *AuthRequest) (string, error) {
	// Check if user not exists
	row := a.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT * FROM users WHERE id = $1)", ar.Login)

	var userExists bool
	if err := row.Scan(&userExists); err != nil {
		logger.Log.Debug(
			"error on scanning row into bool",
			zap.Error(err),
		)
		return "", err
	}

	if err := row.Err(); err != nil {
		logger.Log.Debug(
			"error on scanning row into bool",
			zap.Error(err),
		)
		return "", err
	}

	if userExists {
		logger.Log.Info(
			"trying to registrate already registered user",
			zap.String("userID", ar.Login),
		)
		return "", ErrUserAlreadyExists
	}

	// Calculate pass hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(ar.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Add user to db
	_, err = a.db.ExecContext(ctx, "INSERT INTO users (id, passwordHash) VALUES ($1, $2)", ar.Login, passwordHash)
	if err != nil {
		return "", err
	}

	return a.GenerateToken(ctx, ar)
}

func (a *Auth) LoginUser(ctx context.Context, ar *AuthRequest) (string, error) {

	// Check if user not exists
	userRow := a.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT * FROM users WHERE id = $1)", ar.Login)

	var userExists bool
	if err := userRow.Scan(&userExists); err != nil {
		logger.Log.Debug(
			"error on scanning row into bool",
			zap.Error(err),
		)
		return "", err
	}

	if err := userRow.Err(); err != nil {
		return "", err
	}

	// Calculate pass hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(ar.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Get passwordHash from db
	var realpasswordHash string
	hashRow := a.db.QueryRowContext(ctx, "SELECT passwodHash from users where id = $1", ar.Login)
	if err := hashRow.Scan(&realpasswordHash); err != nil {
		logger.Log.Debug(
			"error on scanning row into string",
			zap.Error(err),
		)
		return "", err
	}

	if !userExists || strings.Compare(string(passwordHash), realpasswordHash) == 0 {
		return "", ErrIncorrectUserCredentials
	}

	return a.GenerateToken(ctx, ar)
}

func (a *Auth) GenerateToken(ctx context.Context, ar *AuthRequest) (string, error) {
	// Return auth token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": ar.Login,
	})

	return token.SignedString(a.key)
}

func (a *Auth) AuthMiddleWare(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authCookie, err := r.Cookie("gophermart-auth")

		if err != nil {
			logger.Log.Debug(
				"unauthorized request",
			)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		token, err := jwt.Parse(authCookie.Value, func(t *jwt.Token) (interface{}, error) {
			return a.key, nil
		})

		if err != nil {
			logger.Log.Debug(
				"cannot parse jwt token passed from client",
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			logger.Log.Debug(
				"unauthorized request",
			)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		logger.Log.Debug(
			"user passed by auth middleware",
			zap.String("userID", claims["userID"].(string)),
		)

		ctx := context.WithValue(r.Context(), "userID", claims["userID"])

		h(w, r.WithContext(ctx))
	})
}
