package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	mockdb "github.com/rdozzi/simple_bank/db/mock"
	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/db/util"
	"github.com/rdozzi/simple_bank/pb"
	"github.com/rdozzi/simple_bank/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
)

func TestUpdateUserAPI(t *testing.T){
	user, _ := randomUser(t)

	newName := util.RandomOwner()
	newEmail := util.RandomEmail()


	testCases := []struct {
		name string
		req *pb.UpdateUserRequest
		buildStubs func(store *mockdb.MockStore)
		buildContext func(t *testing.T,tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req: (&pb.UpdateUserRequest_builder{
				Username: &user.Username,
				FullName: &newName,
				Email: &newEmail,
			}).Build(),
			buildStubs: func(store *mockdb.MockStore){
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: sql.NullString{
						String: newName,
						Valid: true,
					},
					Email: sql.NullString{
						String: newEmail,
						Valid: true,
					},
				}
				updatedUser := db.Users{
					Username: user.Username,
					HashedPassword: user.HashedPassword,
					FullName: newName,
					Email: newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreateAt: user.CreateAt,
					IsEmailVerified: user.IsEmailVerified,
				}
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).Times(1).Return(updatedUser,nil)
			},
			buildContext: func(t *testing.T,tokenMaker token.Maker) context.Context {
				ctx := context.Background()
				accessToken, _, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t,err)
				bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
				md := metadata.MD{
					authorizationHeader: []string{
						bearerToken,
					},
				}
				return metadata.NewIncomingContext(ctx,md)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error){
				require.NoError(t,err)
				require.NotNil(t,res)
				updatedUser := res.GetUser()
				require.Equal(t,user.Username,updatedUser.GetUsername())
				require.Equal(t,newName,updatedUser.GetFullName())
				require.Equal(t,newEmail,updatedUser.GetEmail())
			},
		},
	}

	for i := range testCases{
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T){
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			tc.buildStubs(store)	
			server := newTestServer(t, store, nil)

			ctx := tc.buildContext(t,server.tokenMaker)
			res, err := server.UpdateUser(ctx,tc.req)
			tc.checkResponse(t,res,err)
		})
	}
}