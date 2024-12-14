package handlers

import (
	"github.com/go-chi/chi/v5"
)

func Setup(r *chi.Mux, srv *ServerHandler, hashKey string) {

	// r.Route("/", func(r chi.Router) {
	// 	r.Get("/", middlewares.HmacValidator(hashKey, middlewares.Gzipper(logger.RequestLogger(srv.AllMetrics))))
	// 	r.Get("/ping", middlewares.Gzipper(logger.RequestLogger(srv.Ping)))
	// 	r.Route("/value", func(r chi.Router) {
	// 		r.Post("/", middlewares.HmacValidator(hashKey, middlewares.Gzipper(logger.RequestLogger(srv.GetValueJSON))))
	// 		r.Get("/{type}/{id}", middlewares.Gzipper(logger.RequestLogger(srv.GetValue)))
	// 	})
	// 	r.Post("/updates/", middlewares.HmacValidator(hashKey, middlewares.Gzipper(logger.RequestLogger(srv.UpdatesJSON))))
	// 	r.Route("/update", func(r chi.Router) {
	// 		r.Post("/", middlewares.HmacValidator(hashKey, middlewares.Gzipper(logger.RequestLogger(srv.UpdateJSON))))
	// 		r.Post("/{type}/{id}/{value}", middlewares.Gzipper(logger.RequestLogger(srv.Update)))
	// 	})
	// })
}

type ServerHandler struct {
}

func NewServerHandler() *ServerHandler {
	return &ServerHandler{}
}
