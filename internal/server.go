package app

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rutkin/gofermart/internal/config"
	"github.com/rutkin/gofermart/internal/handlers"
	"github.com/rutkin/gofermart/internal/logger"
	"github.com/rutkin/gofermart/internal/middleware"
	"go.uber.org/zap"
)

func MakeServer(config *config.Config) (*Server, error) {
	handler, err := handlers.NewHandler(config)
	if err != nil {
		return nil, err
	}
	return &Server{config, handler}, nil
}

type Server struct {
	config  *config.Config
	handler *handlers.Handler
}

func (s *Server) Start() {
	logger.Log.Info("running server", zap.String("address", s.config.RunAddress))
	err := http.ListenAndServe(s.config.RunAddress, s.newRouter())
	if err != nil {
		panic(err)
	}
	logger.Log.Info("Server stopped")
}

func (s *Server) newRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/api/user/register", s.handler.Register)
	r.Post("/api/user/login", s.handler.Login)
	userIDRouter := r.With(middleware.WithAuth)
	userIDRouter.Post("/api/user/orders", s.handler.CreateOrder)
	userIDRouter.Get("/api/user/orders", s.handler.GetOrders)
	userIDRouter.Get("/api/user/balance", s.handler.GetBalance)
	return r
}
