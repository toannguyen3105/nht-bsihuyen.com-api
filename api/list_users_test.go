package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/toannguyen3105/nht-bsihuyen.com-api/db/mock"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
	"github.com/toannguyen3105/nht-bsihuyen.com-api/token"
)

func TestListUsersAPI(t *testing.T) {
	user, _ := randomUser(t)
	adminRole := db.Role{
		ID:   1,
		Name: "admin",
	}

	n := 5
	users := make([]db.User, n)
	for i := 0; i < n; i++ {
		users[i] = user
	}

	testCases := []struct {
		name          string
		query         gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: gin.H{
				"page_id":   1,
				"page_size": n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetRolesForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.Role{adminRole}, nil)
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Times(1).
					Return(users, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUsers(t, recorder.Body, users)
			},
		},
		{
			name: "NoAuthorization",
			query: gin.H{
				"page_id":   1,
				"page_size": n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Forbidden",
			query: gin.H{
				"page_id":   1,
				"page_size": n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetRolesForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.Role{}, nil) // No admin role
				store.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: gin.H{
				"page_id":   1,
				"page_size": n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetRolesForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.Role{adminRole}, nil)
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageID",
			query: gin.H{
				"page_id":   -1,
				"page_size": n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetRolesForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.Role{adminRole}, nil)
				store.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/users"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%v", tc.query["page_id"]))
			q.Add("page_size", fmt.Sprintf("%v", tc.query["page_size"]))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchUsers(t *testing.T, body *bytes.Buffer, users []db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var apiResponse APIResponse
	err = json.Unmarshal(data, &apiResponse)
	require.NoError(t, err)

	require.Equal(t, "success", apiResponse.Status)

	var gotUsers []userResponse
	jsonData, err := json.Marshal(apiResponse.Data)
	require.NoError(t, err)
	err = json.Unmarshal(jsonData, &gotUsers)
	require.NoError(t, err)

	require.Equal(t, len(users), len(gotUsers))
	for i, user := range users {
		require.Equal(t, user.Username, gotUsers[i].Username)
		require.Equal(t, user.FullName, gotUsers[i].FullName)
		require.Equal(t, user.Email, gotUsers[i].Email)
	}
}
