package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

func ValidateNumber(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error(
				"error on reading request body",
				zap.Error(err),
			)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		_, err = strconv.ParseInt(string(body), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Debug(
				"passed invalid number",
				zap.Error(err),
			)
			return
		}

		h(w, r)
	}
}

func ValidateJSON(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error(
				"error on reading request body",
				zap.Error(err),
			)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		if !json.Valid(body) {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Debug(
				"passed invalid json",
				zap.Error(err),
			)
			return
		}

		h(w, r)
	}
}
