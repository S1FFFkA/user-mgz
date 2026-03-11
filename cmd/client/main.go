package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	addr := os.Getenv("USER_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		failf("connect grpc: %v", err)
	}
	defer conn.Close()

	client := userv1.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	switch os.Args[1] {
	case "create-user":
		createUser(ctx, client, os.Args[2:])
	case "get-user":
		getUser(ctx, client, os.Args[2:])
	case "update-user":
		updateUser(ctx, client, os.Args[2:])
	case "delete-user":
		deleteUser(ctx, client, os.Args[2:])
	case "list-users":
		listUsers(ctx, client, os.Args[2:])
	case "photo-upload-url":
		photoUploadURL(ctx, client, os.Args[2:])
	case "photo-confirm":
		photoConfirm(ctx, client, os.Args[2:])
	case "photo-download-url":
		photoDownloadURL(ctx, client, os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func createUser(ctx context.Context, c userv1.UserServiceClient, args []string) {
	fs := flag.NewFlagSet("create-user", flag.ExitOnError)
	firstName := fs.String("first-name", "", "required")
	lastName := fs.String("last-name", "", "required")
	email := fs.String("email", "", "required")
	birthDate := fs.String("birth-date", "", "required (YYYY-MM-DD)")
	toiler := fs.Int("toiler", 0, "required 1..10")
	sex := fs.String("sex", "male", "male|female")
	primaryKey := fs.String("primary-key", "", "required S3 object key")
	primaryURL := fs.String("primary-url", "", "required S3 object url/key")
	cityID := fs.Int64("city-id", 0, "optional city id")
	height := fs.Int("height", 0, "optional cm")
	bio := fs.String("bio", "", "optional")
	alco := fs.String("alcohol", "", "optional")
	smoke := fs.String("smoking", "", "optional")
	_ = fs.Parse(args)

	require(*firstName != "" && *lastName != "" && *email != "" && *birthDate != "" &&
		*toiler > 0 && *primaryKey != "" && *primaryURL != "",
		"required: --first-name --last-name --email --birth-date --toiler --primary-key --primary-url")

	req := &userv1.CreateUserRequest{
		FirstName:             *firstName,
		LastName:              *lastName,
		Email:                 *email,
		BirthDate:             *birthDate,
		ToilerScore:           int32(*toiler),
		Sex:                   parseSex(*sex),
		PrimaryPhotoObjectKey: *primaryKey,
		PrimaryPhotoUrl:       *primaryURL,
		Bio:                   *bio,
		AlcoholInfo:           *alco,
		SmokingInfo:           *smoke,
	}
	if *cityID > 0 {
		req.CityId = *cityID
	}
	if *height > 0 {
		req.HeightCm = int32(*height)
	}

	resp, err := c.CreateUser(ctx, req)
	if err != nil {
		failf("CreateUser: %v", err)
	}
	printProto(resp)
}

func getUser(ctx context.Context, c userv1.UserServiceClient, args []string) {
	fs := flag.NewFlagSet("get-user", flag.ExitOnError)
	userID := fs.String("user-id", "", "required UUIDv7")
	_ = fs.Parse(args)
	require(*userID != "", "required: --user-id")

	resp, err := c.GetUser(ctx, &userv1.GetUserRequest{UserId: *userID})
	if err != nil {
		failf("GetUser: %v", err)
	}
	printProto(resp)
}

func updateUser(ctx context.Context, c userv1.UserServiceClient, args []string) {
	fs := flag.NewFlagSet("update-user", flag.ExitOnError)
	userID := fs.String("user-id", "", "required UUIDv7")
	firstName := fs.String("first-name", "", "optional")
	lastName := fs.String("last-name", "", "optional")
	email := fs.String("email", "", "optional")
	birthDate := fs.String("birth-date", "", "optional YYYY-MM-DD")
	toiler := fs.Int("toiler", 0, "optional 1..10")
	sex := fs.String("sex", "", "optional male|female")
	primaryKey := fs.String("primary-key", "", "optional")
	primaryURL := fs.String("primary-url", "", "optional")
	cityID := fs.Int64("city-id", 0, "optional")
	height := fs.Int("height", 0, "optional")
	bio := fs.String("bio", "", "optional")
	alco := fs.String("alcohol", "", "optional")
	smoke := fs.String("smoking", "", "optional")
	_ = fs.Parse(args)
	require(*userID != "", "required: --user-id")

	req := &userv1.UpdateUserRequest{UserId: *userID}
	if *firstName != "" {
		req.FirstName = strPtr(*firstName)
	}
	if *lastName != "" {
		req.LastName = strPtr(*lastName)
	}
	if *email != "" {
		req.Email = strPtr(*email)
	}
	if *birthDate != "" {
		req.BirthDate = strPtr(*birthDate)
	}
	if *toiler > 0 {
		req.ToilerScore = int32Ptr(int32(*toiler))
	}
	if *sex != "" {
		s := parseSex(*sex)
		req.Sex = &s
	}
	if *primaryKey != "" {
		req.PrimaryPhotoObjectKey = strPtr(*primaryKey)
	}
	if *primaryURL != "" {
		req.PrimaryPhotoUrl = strPtr(*primaryURL)
	}
	if *cityID > 0 {
		req.CityId = int64Ptr(*cityID)
	}
	if *height > 0 {
		req.HeightCm = int32Ptr(int32(*height))
	}
	if *bio != "" {
		req.Bio = strPtr(*bio)
	}
	if *alco != "" {
		req.AlcoholInfo = strPtr(*alco)
	}
	if *smoke != "" {
		req.SmokingInfo = strPtr(*smoke)
	}

	resp, err := c.UpdateUser(ctx, req)
	if err != nil {
		failf("UpdateUser: %v", err)
	}
	printProto(resp)
}

func deleteUser(ctx context.Context, c userv1.UserServiceClient, args []string) {
	fs := flag.NewFlagSet("delete-user", flag.ExitOnError)
	userID := fs.String("user-id", "", "required UUIDv7")
	_ = fs.Parse(args)
	require(*userID != "", "required: --user-id")

	resp, err := c.DeleteUser(ctx, &userv1.DeleteUserRequest{UserId: *userID})
	if err != nil {
		failf("DeleteUser: %v", err)
	}
	printProto(resp)
}

func listUsers(ctx context.Context, c userv1.UserServiceClient, args []string) {
	fs := flag.NewFlagSet("list-users", flag.ExitOnError)
	limit := fs.Int("limit", 20, "optional")
	offset := fs.Int("offset", 0, "optional")
	cityID := fs.Int64("city-id", 0, "optional")
	_ = fs.Parse(args)

	resp, err := c.ListUsers(ctx, &userv1.ListUsersRequest{
		Limit:  int32(*limit),
		Offset: int32(*offset),
		CityId: *cityID,
	})
	if err != nil {
		failf("ListUsers: %v", err)
	}
	printProto(resp)
}

func photoUploadURL(ctx context.Context, c userv1.UserServiceClient, args []string) {
	fs := flag.NewFlagSet("photo-upload-url", flag.ExitOnError)
	userID := fs.String("user-id", "", "required UUIDv7")
	photoType := fs.String("type", "primary", "primary|extra")
	extraPos := fs.Int("extra-pos", 0, "required for extra: 1..6")
	contentType := fs.String("content-type", "image/jpeg", "image/jpeg|image/png|image/webp")
	contentLength := fs.Int64("content-length", 1024, "bytes")
	_ = fs.Parse(args)
	require(*userID != "", "required: --user-id")

	req := &userv1.GetUserPhotoUploadUrlRequest{
		UserId:        *userID,
		PhotoType:     parsePhotoType(*photoType),
		ExtraPosition: int32(*extraPos),
		ContentType:   *contentType,
		ContentLength: *contentLength,
	}
	resp, err := c.GetUserPhotoUploadUrl(ctx, req)
	if err != nil {
		failf("GetUserPhotoUploadUrl: %v", err)
	}
	printProto(resp)
}

func photoConfirm(ctx context.Context, c userv1.UserServiceClient, args []string) {
	fs := flag.NewFlagSet("photo-confirm", flag.ExitOnError)
	userID := fs.String("user-id", "", "required UUIDv7")
	photoType := fs.String("type", "primary", "primary|extra")
	extraPos := fs.Int("extra-pos", 0, "required for extra")
	objectKey := fs.String("object-key", "", "required")
	_ = fs.Parse(args)
	require(*userID != "" && *objectKey != "", "required: --user-id --object-key")

	resp, err := c.ConfirmUserPhotoUpload(ctx, &userv1.ConfirmUserPhotoUploadRequest{
		UserId:        *userID,
		PhotoType:     parsePhotoType(*photoType),
		ExtraPosition: int32(*extraPos),
		ObjectKey:     *objectKey,
	})
	if err != nil {
		failf("ConfirmUserPhotoUpload: %v", err)
	}
	printProto(resp)
}

func photoDownloadURL(ctx context.Context, c userv1.UserServiceClient, args []string) {
	fs := flag.NewFlagSet("photo-download-url", flag.ExitOnError)
	userID := fs.String("user-id", "", "required UUIDv7")
	photoType := fs.String("type", "primary", "primary|extra")
	photoID := fs.Int64("photo-id", 0, "required for extra")
	_ = fs.Parse(args)
	require(*userID != "", "required: --user-id")

	resp, err := c.GetUserPhotoDownloadUrl(ctx, &userv1.GetUserPhotoDownloadUrlRequest{
		UserId:    *userID,
		PhotoType: parsePhotoType(*photoType),
		PhotoId:   *photoID,
	})
	if err != nil {
		failf("GetUserPhotoDownloadUrl: %v", err)
	}
	printProto(resp)
}

func parseSex(raw string) userv1.Sex {
	switch raw {
	case "male":
		return userv1.Sex_SEX_MALE
	case "female":
		return userv1.Sex_SEX_FEMALE
	default:
		failf("invalid --sex value: %s (allowed male|female)", raw)
		return userv1.Sex_SEX_UNSPECIFIED
	}
}

func parsePhotoType(raw string) userv1.PhotoType {
	switch raw {
	case "primary":
		return userv1.PhotoType_PHOTO_TYPE_PRIMARY
	case "extra":
		return userv1.PhotoType_PHOTO_TYPE_EXTRA
	default:
		failf("invalid --type value: %s (allowed primary|extra)", raw)
		return userv1.PhotoType_PHOTO_TYPE_UNSPECIFIED
	}
}

func strPtr(v string) *string {
	return &v
}

func int32Ptr(v int32) *int32 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func printProto(v proto.Message) {
	out, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(v)
	if err != nil {
		failf("marshal response: %v", err)
	}
	fmt.Println(string(out))
}

func require(ok bool, msg string) {
	if ok {
		return
	}
	failf(msg)
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func usage() {
	fmt.Print(`Usage:
  go run ./cmd/client <command> [flags]

Commands:
  create-user         --first-name ... --last-name ... --email ... --birth-date YYYY-MM-DD --toiler 1..10 --sex male|female --primary-key ... --primary-url ...
  get-user            --user-id <uuid>
  update-user         --user-id <uuid> [--first-name ... --last-name ... --email ... --birth-date ... --toiler ... --sex ... --primary-key ... --primary-url ...]
  delete-user         --user-id <uuid>
  list-users          [--limit 20 --offset 0 --city-id 0]
  photo-upload-url    --user-id <uuid> --type primary|extra [--extra-pos 1..6] --content-type image/jpeg --content-length 12345
  photo-confirm       --user-id <uuid> --type primary|extra [--extra-pos 1..6] --object-key <s3-object-key>
  photo-download-url  --user-id <uuid> --type primary|extra [--photo-id <id-for-extra>]
`)
}
