package gapi

import (
	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.Users) *pb.User {
	return (&pb.User_builder{
        Username:          &user.Username,
        FullName:          &user.FullName,
        Email:             &user.Email,
        PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
        CreatedAt:         timestamppb.New(user.CreateAt),
    }).Build()
}