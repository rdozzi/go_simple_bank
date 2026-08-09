package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/rdozzi/simple_bank/db/mock"
	db "github.com/rdozzi/simple_bank/db/sqlc"
	"github.com/rdozzi/simple_bank/db/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)
	


func TestGetAccountAPI(t *testing.T){
	account := randomAccount()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)

	// build stubs
	store.EXPECT().GetAccount(gomock.Any(),gomock.Eq(account.ID)).Times(1).Return(account,nil)

	// start test server and send request
	server := NewServer(store)
	recorder := httptest.NewRecorder()

	url := fmt.Sprintf("/accounts/%d", account.ID)
	request, err := http.NewRequest(http.MethodGet,url,nil)
	require.NoError(t,err)

	server.router.ServeHTTP(recorder,request)
	
	// check response
	require.Equal(t,http.StatusOK,recorder.Code)
	requireBodyMatchAccount(t,recorder.Body,account)
	
}

func randomAccount() db.Accounts{
	return db.Accounts{
		ID: util.RandomInt(1,1000),
		Owner: util.RandomOwner(),
		Balance: util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Accounts){
	data, err := io.ReadAll(body)
	require.NoError(t,err)

	var gotAccount db.Accounts
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t,err)
	require.Equal(t,account,gotAccount)

}