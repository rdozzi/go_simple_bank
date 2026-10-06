package gapi

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	mockdb "github.com/rdozzi/simple_bank/db/mock"
	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/db/util"
	"github.com/rdozzi/simple_bank/pb"
	mockwk "github.com/rdozzi/simple_bank/worker/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type eqCreateUserTxParamsMatcher struct {
	arg db.CreateUserTxParams
	password string
}

func (expected eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(expected.password,actualArg.HashedPassword)
	if err != nil {
		return false
	}

	expected.arg.HashedPassword = actualArg.HashedPassword
	if !reflect.DeepEqual(expected.arg.CreateUserParams,actualArg.CreateUserParams) {
		return false
	}

	return true
}

func (e eqCreateUserTxParamsMatcher) String() string{
	return fmt.Sprintf("matches arg %v and password %v",e.arg, e.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams,password string) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg,password}
}

func randomUser(t *testing.T) (user db.Users, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.Users{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func TestCreateUserAPI(t *testing.T){
	user, password := randomUser(t)

	testCases := []struct {
		name string
		req *pb.CreateUserRequest
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: (&pb.CreateUserRequest_builder{
				Username: &user.Username,
				Password: &password,
				FullName: &user.FullName,
				Email: &user.Email,
			}).Build(),
			buildStubs: func(store *mockdb.MockStore){
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email: user.Email,
					},
				}
				store.EXPECT().CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg,password)).Times(1).Return(db.CreateUserTxResult{User: user},nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error){
				require.NoError(t,err)
				require.NotNil(t,res)
				createdUser := res.GetUser()
				require.Equal(t,user.Username,createdUser.GetUsername())
				require.Equal(t,user.FullName,createdUser.GetFullName())
				require.Equal(t,user.Email,createdUser.GetEmail())
			},
		},
	}

	for i := range testCases{
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T){
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			taskDistributor := mockwk.NewMockTaskDistributor(ctrl)

			server := newTestServer(t, store, taskDistributor)
			res, err := server.CreateUser(context.Background(),tc.req)

			tc.checkResponse(t,res,err)
		})
	}
}