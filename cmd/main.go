package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	trmmanager "github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	"github.com/S1FFFkA/user-mgz/internal/config"
	pushtokengrpc "github.com/S1FFFkA/user-mgz/internal/delivery/grpc/pushtoken"
	usergrpc "github.com/S1FFFkA/user-mgz/internal/delivery/grpc/user"
	pushtokenrepo "github.com/S1FFFkA/user-mgz/internal/repository/pushtoken"
	s3repo "github.com/S1FFFkA/user-mgz/internal/repository/s3"
	userrepo "github.com/S1FFFkA/user-mgz/internal/repository/user"
	userphotorepo "github.com/S1FFFkA/user-mgz/internal/repository/userphoto"
	"github.com/S1FFFkA/user-mgz/internal/service"
	userservice "github.com/S1FFFkA/user-mgz/internal/service/user"
	userphotoservice "github.com/S1FFFkA/user-mgz/internal/service/userphoto"
	pgstorage "github.com/S1FFFkA/user-mgz/internal/storage/postgres"
	s3storage "github.com/S1FFFkA/user-mgz/internal/storage/s3"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/grpc/v1"
	"github.com/S1FFFkA/user-mgz/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// S3-репозиторий удовлетворяет контракту сервиса (проверка на этапе компиляции, без импорта service из repository).
var _ service.S3ObjectStorageInterface = (*s3repo.Repository)(nil)

func main() {
	log, err := logger.NewJSON()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = log.Sync()
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgstorage.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to initialize postgres pool", zap.Error(err))
	}
	defer pool.Close()

	s3Client, err := s3storage.NewClient(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3UseSSL)
	if err != nil {
		log.Fatal("failed to initialize s3 client", zap.Error(err))
	}

	usersR := userrepo.New(pool)
	photosR := userphotorepo.New(pool)
	s3R := s3repo.New(s3Client, cfg.S3Bucket)
	txm := service.NewTxManager(trmmanager.Must(trmpgx.NewDefaultFactory(pool)))
	userSvc := userservice.NewService(usersR, photosR, txm, log)
	photoSvc := userphotoservice.NewService(usersR, photosR, s3R, userSvc, log)
	pushTokenR := pushtokenrepo.New(pool)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen",
			zap.String("port", cfg.GRPCPort),
			zap.Error(err),
		)
	}

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Info("metrics server started", zap.String("port", "9102"))

		if err := http.ListenAndServe(":9102", metricsMux); err != nil {
			log.Warn("metrics server stopped", zap.Error(err))
		}
	}()
	
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, usergrpc.NewServer(userSvc, photoSvc, log))
	userv1.RegisterPushTokenServiceServer(grpcServer, pushtokengrpc.NewServer(pushTokenR, log))

	reflection.Register(grpcServer)

	errCh := make(chan error, 1)
	go func() {
		log.Info("user-mgz gRPC server started",
			zap.String("port", cfg.GRPCPort),
			zap.String("s3_endpoint", cfg.S3Endpoint),
			zap.String("s3_bucket", cfg.S3Bucket),
		)
		errCh <- grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received, draining gRPC")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			log.Info("gRPC graceful stop completed")
		case <-shutdownCtx.Done():
			log.Warn("graceful stop timeout, forcing stop")
			grpcServer.Stop()
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}
}
