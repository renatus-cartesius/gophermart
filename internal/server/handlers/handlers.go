package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/renatus-cartesius/gophermart/internal/loyalty"
	"github.com/renatus-cartesius/gophermart/internal/server/middlewares"
	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

func Setup(r *chi.Mux, srv *ServerHandler) {

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Get("/orders", middlewares.Gzipper(logger.RequestLogger(srv.GetOrders)))
			r.Post("/orders", middlewares.Gzipper(logger.RequestLogger(srv.UploadOrder)))
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
}

func NewServerHandler(l *loyalty.Loyalty) *ServerHandler {
	return &ServerHandler{
		l: l,
	}
}

func (s ServerHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := "5c18f4b8-bbb8-11ef-bd1a-8bd0750e0c51"

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
	userID := "5c18f4b8-bbb8-11ef-bd1a-8bd0750e0c51"

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
			)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, loyalty.ErrOrderUploadedAnotherUser) {
			logger.Log.Error(
				"client passed order already uploaded by another user",
				zap.Error(err),
			)
			w.WriteHeader(http.StatusConflict)
			return
		}
		logger.Log.Error(
			"something went wrong when uploading order",
			zap.Error(err),
		)
	}

	w.WriteHeader(http.StatusOK)
}
