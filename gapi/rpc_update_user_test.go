package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	mockdb "github.com/rdozzi/simple_bank/db/mock"
	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/db/util"
	"github.com/rdozzi/simple_bank/pb"
	"github.com/rdozzi/simple_bank/token"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUserAPI(t *testing.T){
	user, _ := randomUser(t)

	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	invalidEmail := "invalid-email"


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
				return newContextWithBearerToken(t,tokenMaker,user.Username,time.Minute)
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
		{
			name: "UserNotFound",
			req: (&pb.UpdateUserRequest_builder{
				Username: &user.Username,
				FullName: &newName,
				Email: &newEmail,
			}).Build(),
			buildStubs: func(store *mockdb.MockStore){
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(1).Return(db.Users{},sql.ErrNoRows)
			},
			buildContext: func(t *testing.T,tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t,tokenMaker,user.Username,time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error){
				require.Error(t,err)
				st, ok := status.FromError(err)
				require.True(t,ok)
				require.Equal(t,codes.NotFound,st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: (&pb.UpdateUserRequest_builder{
				Username: &user.Username,
				FullName: &newName,
				Email: &newEmail,
			}).Build(),
			buildStubs: func(store *mockdb.MockStore){
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			buildContext: func(t *testing.T,tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t,tokenMaker,user.Username,-time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error){
				require.Error(t,err)
				st, ok := status.FromError(err)
				require.True(t,ok)
				require.Equal(t,codes.Unauthenticated,st.Code())
			},
		},
		{
			name: "NoAuthorization",
			req: (&pb.UpdateUserRequest_builder{
				Username: &user.Username,
				FullName: &newName,
				Email: &newEmail,
			}).Build(),
			buildStubs: func(store *mockdb.MockStore){
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			buildContext: func(t *testing.T,tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error){
				require.Error(t,err)
				st, ok := status.FromError(err)
				require.True(t,ok)
				require.Equal(t,codes.Unauthenticated,st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req: (&pb.UpdateUserRequest_builder{
				Username: &user.Username,
				FullName: &newName,
				Email: &invalidEmail,
			}).Build(),
			buildStubs: func(store *mockdb.MockStore){
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			buildContext: func(t *testing.T,tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t,tokenMaker,user.Username,time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error){
				require.Error(t,err)
				st, ok := status.FromError(err)
				require.True(t,ok)
				require.Equal(t,codes.InvalidArgument,st.Code())
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