package auth

import (
	"context"

	ssov1 "github.com/A-PseudoCode-A/proto_files_grpc_sso/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emtpyEmail    = ""
	emtpyPassword = ""
	emptyValue    = 0
)

type Ayth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Ayth
}

func Register(gRPC *grpc.Server, auth Ayth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	//auth из сервисного слоя
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	panic("implement me")
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	panic("implement me")
}

func validateLogin(req *ssov1.LoginRequest) error {
	// Валидация данных (ручная)
	if req.GetEmail() == emtpyEmail {
		return status.Error(codes.InvalidArgument, "invalid argument")
	}

	if req.GetPassword() == emtpyPassword {
		return status.Error(codes.InvalidArgument, "invalid argument")
	}

	if req.AppId == emptyValue {
		return status.Error(codes.InvalidArgument, "invalid argument")
	}

	return nil
}
