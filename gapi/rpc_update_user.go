package gapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/db/util"
	"github.com/rdozzi/simple_bank/pb"
	"github.com/rdozzi/simple_bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {

	authPayload, err := server.authorizeUser(ctx, []string{util.BankerRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, InvalidArgumentError(violations)
	}

	if authPayload.Role != util.BankerRole && authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.HasFullName(),
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.HasEmail(),
		},

	}

	if req.HasPassword() {
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
		}

		arg.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid: true,
		}

		arg.PasswordChangedAt = pgtype.Timestamptz{
			Time: time.Now(),
			Valid: true,
		}
	}

	user, err := server.store.UpdateUser(ctx,arg)
	if err != nil {
		if errors.Is(err,db.ErrRecordNotFound){
			return nil, status.Errorf(codes.NotFound,"user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	rsp := (&pb.UpdateUserResponse_builder{
		User: convertUser(user),
	}).Build()
	
	return rsp, nil

}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation){

	if err := val.ValidateUsername(req.GetUsername()); err != nil{
		violations = append(violations, fieldViolation("username",err))
	}

	if req.HasPassword() {
		if err := val.ValidatePassword(req.GetPassword()); err != nil{
			violations = append(violations, fieldViolation("password",err))
		}
	}

	if req.HasFullName() {
		if err := val.ValidateFullName(req.GetFullName()); err != nil{
			violations = append(violations, fieldViolation("full_name",err))
		}
	}

	if req.HasEmail(){
		if err := val.ValidateEmail(req.GetEmail()); err != nil{
			violations = append(violations, fieldViolation("email",err))
		}
	}

	return violations
}