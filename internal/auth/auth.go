package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

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
	RegisterUser(ctx context.Context, ar *AuthRequest) (*http.Cookie, error)
	LoginUser(ctx context.Context, ar *AuthRequest) (*http.Cookie, error)
	AuthMiddleWare(h http.HandlerFunc) http.HandlerFunc
}

type AuthStorager interface {
	IsUserExists(ctx context.Context, userID string) (bool, error)
	AddUser(ctx context.Context, userID, passwordHash string) error
	GetHash(ctx context.Context, userID string) (string, error)
}

type Auth struct {
	key     []byte
	storage AuthStorager
}

func NewAuth(key []byte, storage AuthStorager) *Auth {

	return &Auth{
		key:     key,
		storage: storage,
	}
}

func (a *Auth) RegisterUser(ctx context.Context, ar *AuthRequest) (*http.Cookie, error) {
	// Check if user not exists
	userExists, err := a.storage.IsUserExists(ctx, ar.Login)
	if err != nil {
		return nil, err
	}

	if userExists {
		logger.Log.Info(
			"trying to registrate already registered user",
			zap.String("userID", ar.Login),
		)
		return nil, ErrUserAlreadyExists
	}

	// Calculate pass hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(ar.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Add user to db
	if err := a.storage.AddUser(ctx, ar.Login, string(passwordHash)); err != nil {
		return nil, err
	}

	return a.GenerateToken(ctx, ar)
}

func (a *Auth) LoginUser(ctx context.Context, ar *AuthRequest) (*http.Cookie, error) {

	// Check if user not exists
	userExists, err := a.storage.IsUserExists(ctx, ar.Login)
	if err != nil {
		return nil, err
	}

	// Get passwordHash from db
	realpasswordHash, err := a.storage.GetHash(ctx, ar.Login)
	if err != nil {
		return nil, err
	}

	// if !userExists || strings.Compare(string(passwordHash), realpasswordHash) == 0 {
	if err := bcrypt.CompareHashAndPassword([]byte(realpasswordHash), []byte(ar.Password)); err != nil || !userExists {
		logger.Log.Debug(
			"incorrect password",
			zap.Error(err),
		)
		return nil, ErrIncorrectUserCredentials
	}

	return a.GenerateToken(ctx, ar)
}

func (a *Auth) GenerateToken(ctx context.Context, ar *AuthRequest) (*http.Cookie, error) {
	expires := time.Now().Add(30 * 24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":  ar.Login,
		"expires": expires,
	})

	tokenString, err := token.SignedString(a.key)
	if err != nil {
		return nil, err
	}

	authCookie := &http.Cookie{
		Name:    "gophermart-auth",
		Value:   tokenString,
		Expires: expires,
	}

	return authCookie, nil
}

func (a *Auth) AuthMiddleWare(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authCookie, err := r.Cookie("gophermart-auth")

		if err != nil {
			logger.Log.Debug(
				"unauthorized request",
			)
			w.WriteHeader(http.StatusUnauthorized)
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
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expire, err := time.Parse(time.RFC3339Nano, claims["expires"].(string))
		if err != nil {
			logger.Log.Debug(
				"error when parsing expire in token",
				zap.Error(err),
			)
		}

		now := time.Now()
		if now.After(expire) {
			logger.Log.Debug(
				"passed outdated token",
				zap.Time("expire", expire),
				zap.Time("now", now),
			)
			w.WriteHeader(http.StatusUnauthorized)
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
