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
	"github.com/toannguyen3105/nht-bsihuyen.com-api/utils"
)

func TestCreatePermissionAPI(t *testing.T) {
	user, _ := randomUser(t)
	permission := randomPermission()

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
				"name":        permission.Name,
				"description": permission.Description.String,
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
					Return([]string{"VIEW_SCREEN_PERMISSION"}, nil)
				arg := db.CreatePermissionParams{
					Name:        permission.Name,
					Description: permission.Description,
				}
				store.EXPECT().
					CreatePermission(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(permission, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPermission(t, recorder.Body, permission)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"name":        permission.Name,
				"description": permission.Description.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreatePermission(gomock.Any(), gomock.Any()).
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

			url := "/permissions"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetPermissionAPI(t *testing.T) {
	user, _ := randomUser(t)
	permission := randomPermission()

	testCases := []struct {
		name          string
		permissionID  int32
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:         "OK",
			permissionID: permission.ID,
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
					Return([]string{"VIEW_SCREEN_PERMISSION"}, nil)
				store.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(permission.ID)).
					Times(1).
					Return(permission, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPermission(t, recorder.Body, permission)
			},
		},
		{
			name:         "NotFound",
			permissionID: permission.ID,
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
					Return([]string{"VIEW_SCREEN_PERMISSION"}, nil)
				store.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(permission.ID)).
					Times(1).
					Return(db.Permission{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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

			url := fmt.Sprintf("/permissions/%d", tc.permissionID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListPermissionsAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 10
	permissions := make([]db.Permission, n)
	for i := 0; i < n; i++ {
		permissions[i] = randomPermission()
	}

	testCases := []struct {
		name          string
		query         string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
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
					Return([]string{"VIEW_SCREEN_PERMISSION"}, nil)
				arg := db.ListPermissionsParams{
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListPermissions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(permissions, nil)
				store.EXPECT().
					CountPermissions(gomock.Any()).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
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

			url := "/permissions" + tc.query
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomPermission() db.Permission {
	return db.Permission{
		ID:   int32(utils.RandomInt(1, 1000)),
		Name: utils.RandomString(10),
		Description: sql.NullString{
			String: utils.RandomString(20),
			Valid:  true,
		},
	}
}

func requireBodyMatchPermission(t *testing.T, body *bytes.Buffer, permission db.Permission) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var apiResponse APIResponse
	err = json.Unmarshal(data, &apiResponse)
	require.NoError(t, err)
	require.Equal(t, "success", apiResponse.Status)

	var gotPermission permissionResponse
	jsonData, err := json.Marshal(apiResponse.Data)
	require.NoError(t, err)
	err = json.Unmarshal(jsonData, &gotPermission)
	require.NoError(t, err)

	require.Equal(t, permission.Name, gotPermission.Name)
	require.Equal(t, permission.Description.String, gotPermission.Description)
}
