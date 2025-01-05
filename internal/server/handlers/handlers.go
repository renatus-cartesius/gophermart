package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/renatus-cartesius/gophermart/internal/accrual"
	"github.com/renatus-cartesius/gophermart/internal/auth"
	"github.com/renatus-cartesius/gophermart/internal/loyalty"
	"github.com/renatus-cartesius/gophermart/internal/server/middlewares"
	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

func Setup(r *chi.Mux, srv *ServerHandler) {

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Get("/orders", srv.a.AuthMiddleWare(middlewares.Gzipper(logger.RequestLogger(srv.GetOrders))))
			r.Post("/orders", srv.a.AuthMiddleWare(middlewares.Gzipper(logger.RequestLogger(srv.UploadOrder))))
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", srv.a.AuthMiddleWare(middlewares.Gzipper(logger.RequestLogger(srv.GetBalance))))
				r.Post("/withdraw", srv.a.AuthMiddleWare(middlewares.Gzipper(logger.RequestLogger(srv.Withdraw))))
			})
			r.Post("/register", middlewares.Gzipper(logger.RequestLogger(srv.RegisterUser)))
			r.Post("/login", middlewares.Gzipper(logger.RequestLogger(srv.LoginUser)))
		})
		// 	r.Get("/ping", middlewares.Gzipper(logger.RequestLogger(srv.Ping)))
		// 	r.Post("/updates/", middlewares.HmacValidator(hashKey, middlewares.Gzipper(logger.RequestLogger(srv.UpdatesJSON))))
		// 	r.Route("/update", func(r chi.Router) {
		// 		r.Post("/", middlewares.HmacValidator(hashKey, middlewares.Gzipper(logger.RequestLogger(srv.UpdateJSON))))
		// 		r.Post("/{type}/{id}/{value}", middlewares.Gzipper(logger.RequestLogger(srv.Update)))
		// 	})
	})
}

type ServerHandler struct {
	l *loyalty.Loyalty
	a auth.Auther
}

func NewServerHandler(l *loyalty.Loyalty, a auth.Auther) *ServerHandler {
	return &ServerHandler{
		l: l,
		a: a,
	}
}

func (s ServerHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	ar := &auth.AuthRequest{}
	if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
		logger.Log.Error(
			"error on unmarshalling auth request body",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authCookie, err := s.a.RegisterUser(r.Context(), ar)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			logger.Log.Error(
				"trying to register user with already registered login",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusConflict)
			return
		} else {
			logger.Log.Error(
				"error when registering user",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, authCookie)
	w.WriteHeader(http.StatusOK)
}

func (s ServerHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	ar := &auth.AuthRequest{}
	if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
		logger.Log.Error(
			"error on unmarshalling auth request body",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authCookie, err := s.a.LoginUser(r.Context(), ar)
	if err != nil {
		if errors.Is(err, auth.ErrIncorrectUserCredentials) {
			logger.Log.Error(
				"trying to login user with invalid credentials",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			logger.Log.Error(
				"error when login user",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, authCookie)
	w.WriteHeader(http.StatusOK)
}

func (s ServerHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var buf bytes.Buffer

	orders, err := s.l.GetOrders(r.Context(), userID)
	if err != nil {
		logger.Log.Error(
			"error on getting orders from loyalty storage",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(&buf).Encode(orders); err != nil {
		logger.Log.Error(
			"error on marshalling orders for user",
			zap.String("userID", userID),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (s ServerHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Error(
			"error on reading request body",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	order, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		logger.Log.Error(
			"error on parsing order number",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = s.l.UploadOrder(r.Context(), userID, order); err != nil {
		if errors.Is(err, loyalty.ErrOrderInvalid) {
			logger.Log.Error(
				"client passed invalid order number",
				zap.Error(err),
				zap.String("userID", userID),
			)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, loyalty.ErrOrderUploadedAnotherUser) {
			logger.Log.Error(
				"client passed order already uploaded by another user",
				zap.Error(err),
				zap.String("userID", userID),
			)
			w.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, loyalty.ErrOrderAlreadyUploaded) {
			logger.Log.Debug(
				"client passed order already by this user",
				zap.String("userID", userID),
			)
			w.WriteHeader(http.StatusOK)
			return
		}
		logger.Log.Error(
			"something went wrong when uploading order",
			zap.Error(err),
		)
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s ServerHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	balance, err := s.l.GetBalance(r.Context(), userID)
	if err != nil {
		logger.Log.Error(
			"error when getting balance",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(balance); err != nil {
		logger.Log.Error(
			"error when marshalling balance",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (s ServerHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	withdrawRequest := &loyalty.Withdraw{}

	if err := json.NewDecoder(r.Body).Decode(&withdrawRequest); err != nil {
		logger.Log.Error(
			"error on unmarshalling withdrawRequest body",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	withdrawRequest.UserID = userID

	if err := s.l.Withdraw(r.Context(), withdrawRequest); err != nil {
		if errors.Is(err, loyalty.ErrWithdrawNotEnoughPoints) {
			w.WriteHeader(http.StatusPaymentRequired)
		}
		if errors.Is(err, accrual.ErrOrderNotProcessed) || errors.Is(err, accrual.ErrOrderNotFound) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		logger.Log.Error(
			"error when making withdraw",
			zap.String("userID", userID),
			zap.Int64("orderID", withdrawRequest.OrderID),
			zap.Error(err),
		)
		return
	}

	w.WriteHeader(http.StatusOK)
}
