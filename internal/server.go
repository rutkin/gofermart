package app

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rutkin/gofermart/internal/handlers"
	"github.com/rutkin/gofermart/internal/logger"
	"go.uber.org/zap"
)

func MakeServer() *Server {
	return &Server{handlers.NewHandler()}
}

type Server struct {
	handler *handlers.Handler
}

func (s *Server) Start(address string) {
	logger.Log.Info("running server", zap.String("address", address))
	err := http.ListenAndServe(address, s.newRouter())
	if err != nil {
		panic(err)
	}
	logger.Log.Info("Server stopped")
}

func (s *Server) newRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/api/user/register", s.handler.Register)
	r.Post("/api/user/login", s.handler.Login)
	return r
}
