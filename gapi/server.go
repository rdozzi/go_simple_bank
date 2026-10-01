package gapi

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/db/util"
	"github.com/rdozzi/simple_bank/pb"
	"github.com/rdozzi/simple_bank/token"
	"github.com/rdozzi/simple_bank/worker"
)

// Server serves gRPC requests for our banking service
type Server struct {
	pb.UnimplementedSimpleBankServer
	config util.Config
	store db.Store
	tokenMaker token.Maker
	router *gin.Engine
	taskDistributor worker.TaskDistributor
}

// NewServer creates a new gRPC server.
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	// tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config: config,
		store: store, 
		tokenMaker: tokenMaker,
		taskDistributor: taskDistributor,
	}
	 
	return server, nil
}