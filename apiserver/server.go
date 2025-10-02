package apiserver

import (
	"context"
	"go-sqs/config"
	"go-sqs/store"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type ApiServer struct {
	Config *config.Config
	logger *slog.Logger
	store  *store.Store
}

func New(config *config.Config, logger *slog.Logger, store *store.Store) *ApiServer {
	return &ApiServer{
		Config: config,
		logger: logger,
		store:  store,
	}
}

func (s *ApiServer) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (s *ApiServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", s.ping)
	mux.HandleFunc("POST /auth/singup", s.signupHandler())

	middleware := NewLoggerMiddleware(s.logger)
	server := &http.Server{
		Addr:    net.JoinHostPort(s.Config.ApiServerHost, s.Config.ApiServerPort),
		Handler: middleware(mux),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("api server failed")
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {

		}
	}()
	wg.Wait()

	return nil
}
