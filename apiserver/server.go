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

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type ApiServer struct {
	Config *config.Config
	logger *slog.Logger
	store  *store.Store
	JwtManager *JwtManager
	sqsClient *sqs.Client
	presignClient *s3.PresignClient
}

func New(config *config.Config, logger *slog.Logger, store *store.Store, jwtManager *JwtManager, sqsClient *sqs.Client, presignClient *s3.PresignClient) *ApiServer {
	return &ApiServer{
		Config: config,
		logger: logger,
		store:  store,
		JwtManager: jwtManager,
		sqsClient: sqsClient,
		presignClient: presignClient,
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
	mux.HandleFunc("POST /auth/singin", s.signinHandler())
	mux.HandleFunc("POST /auth/refresh", s.tokenRefreshHandler())
	mux.HandleFunc("POST /reports", s.createReportHandler())
	mux.HandleFunc("GET /reports/{id}", s.getReportHandler())

	middleware := NewLoggerMiddleware(s.logger)
	middleware = NewAuthMiddleware(s.JwtManager, s.store.Users)
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
