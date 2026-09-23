package gapi

import (
	"context"

	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/db/util"
	"github.com/rdozzi/simple_bank/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to find user: %s", err)
	}

	err = util.CheckPassword(req.GetPassword(),user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "incorrect password: %s", err)
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token: %s", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token: %s", err)
	}

	mtdt := server.extractMetadata(ctx)
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
	ID: refreshPayload.ID,
	Username: user.Username,
	RefreshToken: refreshToken,
	UserAgent: mtdt.UserAgent,
	ClientIp: mtdt.ClientIP,
	IsBlocked: false,
	ExpiresAt: refreshPayload.ExpiresAt.Time,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %s", err)
	}

	sessionID := session.ID.String()

	rsp := (&pb.LoginUserResponse_builder{
		User:      convertUser(user),
		SessionId: &sessionID,
		AccessToken: &accessToken,
		RefreshToken: &refreshToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpiresAt.Time),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiresAt.Time),
	}).Build()

	return rsp, nil
}
