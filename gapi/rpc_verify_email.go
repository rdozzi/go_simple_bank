package gapi

import (
	"context"

	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/pb"
	"github.com/rdozzi/simple_bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {

	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, InvalidArgumentError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(ctx,db.VerifyEmailTxParams{
		EmailId: req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal,"failed to verify email")
	}

	rsp := (&pb.VerifyEmailResponse_builder{
		IsVerified: &txResult.User.IsEmailVerified,
	}).Build()
	
	return rsp, nil

}

func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation){

	if err := val.ValidateEmailId(req.GetEmailId()); err != nil{
		violations = append(violations, fieldViolation("email_id",err))
	}

	if err := val.ValidateSecretCode(req.GetSecretCode()); err != nil{
		violations = append(violations, fieldViolation("secret_code",err))
	}

	return violations
}