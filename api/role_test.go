package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	mockdb "github.com/toannguyen3105/nht-bsihuyen.com-api/db/mock"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
	"github.com/toannguyen3105/nht-bsihuyen.com-api/token"
	"github.com/toannguyen3105/nht-bsihuyen.com-api/utils"
)

func TestCreateRoleAPI(t *testing.T) {
	user, _ := randomUser(t)
	role := randomRole()

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"name":        role.Name,
				"description": role.Description.String,
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
					GetPermissionsForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]string{"VIEW_SCREEN_ROLE"}, nil)
				arg := db.CreateRoleParams{
					Name:        role.Name,
					Description: role.Description,
				}
				store.EXPECT().
					CreateRole(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(role, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchRole(t, recorder.Body, role)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"name":        role.Name,
				"description": role.Description.String,
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
					GetPermissionsForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]string{"VIEW_SCREEN_ROLE"}, nil)
				store.EXPECT().
					CreateRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Role{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidData",
			body: gin.H{
				"name": "", // Invalid name
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
					GetPermissionsForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]string{"VIEW_SCREEN_ROLE"}, nil)
				store.EXPECT().
					CreateRole(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DuplicateRole",
			body: gin.H{
				"name":        role.Name,
				"description": role.Description.String,
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
					GetPermissionsForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]string{"VIEW_SCREEN_ROLE"}, nil)
				store.EXPECT().
					CreateRole(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Role{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// No auth setup
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateRole(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/roles"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomRole() db.Role {
	return db.Role{
		ID:   int32(utils.RandomInt(1, 1000)),
		Name: utils.RandomString(6),
		Description: sql.NullString{
			String: utils.RandomString(20),
			Valid:  true,
		},
	}
}

func TestListRoles(t *testing.T) {
	user, _ := randomUser(t)
	n := 10
	roles := make([]db.Role, n)
	for i := 0; i < n; i++ {
		roles[i] = randomRole()
	}

	testCases := []struct {
		name          string
		query         string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "?page_id=1&page_size=10",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetPermissionsForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]string{"VIEW_SCREEN_ROLE"}, nil)
				arg := db.ListRolesParams{
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListRoles(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(roles, nil)
				store.EXPECT().
					CountRoles(gomock.Any()).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "NoAuthorization",
			query: "?page_id=1&page_size=10",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListRoles(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:  "UnauthorizedUser",
			query: "?page_id=1&page_size=10",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetPermissionsForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]string{"VIEW_SCREEN_USER"}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:  "InternalError",
			query: "?page_id=1&page_size=10",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetPermissionsForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]string{"VIEW_SCREEN_ROLE"}, nil)
				store.EXPECT().
					ListRoles(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Role{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidPageID",
			query: "?page_id=-1&page_size=10",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					GetPermissionsForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]string{"VIEW_SCREEN_ROLE"}, nil)
				store.EXPECT().
					ListRoles(gomock.Any(), gomock.Any()).
					Times(0)
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

			url := "/roles" + tc.query
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchRole(t *testing.T, body *bytes.Buffer, role db.Role) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var apiResponse APIResponse
	err = json.Unmarshal(data, &apiResponse)
	require.NoError(t, err)
	require.Equal(t, "success", apiResponse.Status)

	var gotRole roleResponse
	jsonData, err := json.Marshal(apiResponse.Data)
	require.NoError(t, err)
	err = json.Unmarshal(jsonData, &gotRole)
	require.NoError(t, err)

	require.Equal(t, role.Name, gotRole.Name)
	require.Equal(t, role.Description.String, gotRole.Description)
}
