package main

import (
	"context"
	"net"
	"net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/S1FFFkA/user-mgz/internal/config"
	s3repo "github.com/S1FFFkA/user-mgz/internal/repository/s3"
	userrepo "github.com/S1FFFkA/user-mgz/internal/repository/user"
	userphotorepo "github.com/S1FFFkA/user-mgz/internal/repository/userphoto"
	pgstorage "github.com/S1FFFkA/user-mgz/internal/storage/postgres"
	s3storage "github.com/S1FFFkA/user-mgz/internal/storage/s3"
	grpcmw "github.com/S1FFFkA/user-mgz/internal/transport/grpc/middleware"
	usertgrpc "github.com/S1FFFkA/user-mgz/internal/transport/grpc/user"
	usersvc "github.com/S1FFFkA/user-mgz/internal/usecase/user"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"github.com/S1FFFkA/user-mgz/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

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

	ctx := context.Background()

	pool, err := pgstorage.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to initialize postgres pool", zap.Error(err))
	}
	defer pool.Close()

	s3Client, err := s3storage.NewClient(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3UseSSL)
	if err != nil {
		log.Fatal("failed to initialize s3 client", zap.Error(err))
	}

	userRepo := userrepo.NewRepository(pool)
	userPhotoRepo := userphotorepo.NewRepository(pool)
	s3Repo := s3repo.NewRepository(s3Client, cfg.S3Bucket)
	userService := usersvc.NewService(userRepo, userPhotoRepo, s3Repo)

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
    	if err := http.ListenAndServe(":9102", metricsMux); err != nil {
        	log.Warn("metrics server stopped", zap.Error(err))
    	}
	}()
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcmw.UnaryTraceInterceptor()),
	)
	userv1.RegisterUserServiceServer(grpcServer, usertgrpc.NewServer(userService, log))

	reflection.Register(grpcServer)

	log.Info("user-mgz gRPC server started",
		zap.String("port", cfg.GRPCPort),
		zap.String("s3_endpoint", cfg.S3Endpoint),
		zap.String("s3_bucket", cfg.S3Bucket),
	)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve gRPC", zap.Error(err))
	}
}
