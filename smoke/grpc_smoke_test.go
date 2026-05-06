//go:build smoke

package smoke

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	trmmanager "github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	"github.com/S1FFFkA/user-mgz/integration/testutil"
	usergrpc "github.com/S1FFFkA/user-mgz/internal/delivery/grpc/user"
	s3repo "github.com/S1FFFkA/user-mgz/internal/repository/s3"
	userrepo "github.com/S1FFFkA/user-mgz/internal/repository/user"
	userphotorepo "github.com/S1FFFkA/user-mgz/internal/repository/userphoto"
	"github.com/S1FFFkA/user-mgz/internal/service"
	userservice "github.com/S1FFFkA/user-mgz/internal/service/user"
	userphotoservice "github.com/S1FFFkA/user-mgz/internal/service/userphoto"
	s3storage "github.com/S1FFFkA/user-mgz/internal/storage/s3"
	userv1 "github.com/S1FFFkA/user-mgz/pkg/grpc/v1"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPC_Smoke_CreateGetListDelete(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("smoke: задайте DATABASE_URL (например postgres из docker compose на localhost:5433)")
	}

	ctx := context.Background()
	log := zap.NewNop()

	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	// gRPC слой обогащает фотки пресайн-ссылками, поэтому нужен живой S3/MinIO.
	// Дефолты совпадают с docker-compose.yml.
	endpoint := getenv("S3_ENDPOINT", "localhost:9000")
	accessKey := getenv("S3_ACCESS_KEY", "minioadmin")
	secretKey := getenv("S3_SECRET_KEY", "minioadmin")
	bucket := getenv("S3_BUCKET", "fotos")
	minioClient, err := s3storage.NewClient(endpoint, accessKey, secretKey, false)
	if err != nil {
		t.Fatalf("create minio client: %v", err)
	}
	s3 := s3repo.New(minioClient, bucket)

	usersR := userrepo.New(env.Pool)
	photosR := userphotorepo.New(env.Pool)
	txm := service.NewTxManager(trmmanager.Must(trmpgx.NewDefaultFactory(env.Pool)))
	userSvc := userservice.NewService(usersR, photosR, txm, log)
	photoSvc := userphotoservice.NewService(usersR, photosR, s3, userSvc, log)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	userv1.RegisterUserServiceServer(srv, usergrpc.NewServer(userSvc, photoSvc, log))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop(); _ = lis.Close() })

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	c := userv1.NewUserServiceClient(conn)

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	created, err := c.CreateUser(reqCtx, &userv1.CreateUserRequest{
		FirstName:             "A",
		LastName:              "B",
		Email:                 "a@b.com",
		BirthDate:             "2000-01-01",
		ToilerScore:           7,
		Sex:                   userv1.Sex_SEX_MALE,
		PrimaryPhotoObjectKey: "k",
		PrimaryPhotoUrl:       "u",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.GetUser().GetId() == "" {
		t.Fatalf("expected id")
	}

	_, err = c.GetUser(reqCtx, &userv1.GetUserRequest{UserId: created.GetUser().GetId()})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}

	list, err := c.ListUsers(reqCtx, &userv1.ListUsersRequest{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(list.GetUsers()) == 0 {
		t.Fatalf("expected users")
	}

	_, err = c.DeleteUser(reqCtx, &userv1.DeleteUserRequest{UserId: created.GetUser().GetId()})
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
}

func TestGRPC_Smoke_UploadDownloadPrimaryPhoto_RoundTripBytes(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("smoke: задайте DATABASE_URL (например postgres из docker compose на localhost:5433)")
	}

	ctx := context.Background()
	log := zap.NewNop()

	env, err := testutil.StartPostgres(ctx, "../migrations")
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = env.Teardown(ctx) }()

	endpoint := getenv("S3_ENDPOINT", "localhost:9000")
	accessKey := getenv("S3_ACCESS_KEY", "minioadmin")
	secretKey := getenv("S3_SECRET_KEY", "minioadmin")
	bucket := getenv("S3_BUCKET", "fotos")
	minioClient, err := s3storage.NewClient(endpoint, accessKey, secretKey, false)
	if err != nil {
		t.Fatalf("create minio client: %v", err)
	}
	requireMinIOForSmoke(t, minioClient, endpoint, bucket)

	s3 := s3repo.New(minioClient, bucket)

	usersR := userrepo.New(env.Pool)
	photosR := userphotorepo.New(env.Pool)
	txm := service.NewTxManager(trmmanager.Must(trmpgx.NewDefaultFactory(env.Pool)))
	userSvc := userservice.NewService(usersR, photosR, txm, log)
	photoSvc := userphotoservice.NewService(usersR, photosR, s3, userSvc, log)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	userv1.RegisterUserServiceServer(srv, usergrpc.NewServer(userSvc, photoSvc, log))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop(); _ = lis.Close() })

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	c := userv1.NewUserServiceClient(conn)

	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	created, err := c.CreateUser(reqCtx, &userv1.CreateUserRequest{
		FirstName:             "A",
		LastName:              "B",
		Email:                 uuid.Must(uuid.NewV7()).String() + "@b.com",
		BirthDate:             "2000-01-01",
		ToilerScore:           7,
		Sex:                   userv1.Sex_SEX_MALE,
		PrimaryPhotoObjectKey: "placeholder.png",
		PrimaryPhotoUrl:       "placeholder",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	userID := created.GetUser().GetId()
	if userID == "" {
		t.Fatalf("expected user id")
	}

	// Small valid 1x1 PNG.
	photoBytes := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89,
		0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54,
		0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05,
		0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
		0xAE, 0x42, 0x60, 0x82,
	}

	up, err := c.GetUserPhotoUploadUrl(reqCtx, &userv1.GetUserPhotoUploadUrlRequest{
		UserId:        userID,
		PhotoType:     userv1.PhotoType_PHOTO_TYPE_PRIMARY,
		ContentType:   "image/png",
		ContentLength: int64(len(photoBytes)),
	})
	if err != nil {
		t.Fatalf("GetUserPhotoUploadUrl: %v", err)
	}
	if up.GetUploadUrl() == "" || up.GetObjectKey() == "" {
		t.Fatalf("expected upload url and object key")
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	putReq, err := http.NewRequestWithContext(reqCtx, http.MethodPut, up.GetUploadUrl(), bytes.NewReader(photoBytes))
	if err != nil {
		t.Fatalf("new put request: %v", err)
	}
	putResp, err := httpClient.Do(putReq)
	if err != nil {
		t.Fatalf("put to upload url: %v", err)
	}
	_ = putResp.Body.Close()
	if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
		t.Fatalf("put failed: %s", putResp.Status)
	}

	_, err = c.ConfirmUserPhotoUpload(reqCtx, &userv1.ConfirmUserPhotoUploadRequest{
		UserId:    userID,
		PhotoType: userv1.PhotoType_PHOTO_TYPE_PRIMARY,
		ObjectKey: up.GetObjectKey(),
	})
	if err != nil {
		t.Fatalf("ConfirmUserPhotoUpload: %v", err)
	}

	down, err := c.GetUserPhotoDownloadUrl(reqCtx, &userv1.GetUserPhotoDownloadUrlRequest{
		UserId:    userID,
		PhotoType: userv1.PhotoType_PHOTO_TYPE_PRIMARY,
	})
	if err != nil {
		t.Fatalf("GetUserPhotoDownloadUrl: %v", err)
	}
	if down.GetDownloadUrl() == "" {
		t.Fatalf("expected download url")
	}

	getResp, err := httpClient.Get(down.GetDownloadUrl())
	if err != nil {
		t.Fatalf("get from download url: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(getResp.Body, 2048))
		t.Fatalf("get failed: %s (%s)", getResp.Status, string(b))
	}
	gotBytes, err := io.ReadAll(getResp.Body)
	if err != nil {
		t.Fatalf("read downloaded bytes: %v", err)
	}
	if !bytes.Equal(gotBytes, photoBytes) {
		t.Fatalf("downloaded bytes mismatch: got=%d want=%d", len(gotBytes), len(photoBytes))
	}

	// cleanup
	_, _ = c.DeleteUser(reqCtx, &userv1.DeleteUserRequest{UserId: userID})
}

func getenv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}

// requireMinIOForSmoke проверяет, что до MinIO (или совместимого S3) реально достучаться и есть bucket —
// иначе GetUserPhotoUploadUrl вернёт gRPC Unavailable («service unavailable»), что выглядит как ошибка сервера.
func requireMinIOForSmoke(t *testing.T, client *minio.Client, endpoint, bucket string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		msg := fmt.Sprintf(
			"MinIO недоступен или неверные ключи (endpoint=%q bucket=%q): %v\n"+
				"Подните MinIO с bucket: docker compose --profile local-s3 up -d minio minio-init\n"+
				"Или переопределите S3_ENDPOINT/S3_ACCESS_KEY/S3_SECRET_KEY/S3_BUCKET.",
			endpoint, bucket, err)
		if os.Getenv("SMOKE_REQUIRE_MINIO") == "1" {
			t.Fatal(msg)
		}
		t.Skip(msg)
	}
	if !exists {
		msg := fmt.Sprintf(
			"Bucket %q на %q не найден (создайте через mc или minio-init в compose).\n"+
				"docker compose --profile local-s3 up -d minio minio-init",
			bucket, endpoint)
		if os.Getenv("SMOKE_REQUIRE_MINIO") == "1" {
			t.Fatal(msg)
		}
		t.Skip(msg)
	}
}
