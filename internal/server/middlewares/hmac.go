package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

func HmacValidator(key string, h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("HashSHA256") == "" {
			h.ServeHTTP(w, r)
			return
		}

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

		sum, err := base64.StdEncoding.DecodeString(r.Header.Get("HashSHA256"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error(
				"error on decoding base64 sha256 hash sum",
				zap.Error(err),
			)
			return
		}

		hash := hmac.New(sha256.New, []byte(key))
		hash.Write(body)

		if !hmac.Equal(sum, hash.Sum(nil)) {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Error(
				"captured invalid sha256 sum",
				zap.Error(err),
				zap.String("reqSum", r.Header.Get("HashSHA256")),
				zap.String("hashSUm", base64.StdEncoding.EncodeToString(hash.Sum(nil))),
			)
			return
		}
		h.ServeHTTP(w, r)
	})
}
