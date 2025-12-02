package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestCreateRolePermissionAPI(t *testing.T) {
	user, _ := randomUser(t)
	rolePermission := randomRolePermission()

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
				"role_id":       rolePermission.RoleID,
				"permission_id": rolePermission.PermissionID,
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
					Return([]string{"VIEW_SCREEN_ROLE_PERMISSION"}, nil)
				arg := db.CreateRolePermissionParams{
					RoleID:       rolePermission.RoleID,
					PermissionID: rolePermission.PermissionID,
				}
				store.EXPECT().
					CreateRolePermission(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(rolePermission, nil)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/role_permissions"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteRolePermissionAPI(t *testing.T) {
	user, _ := randomUser(t)
	rolePermission := randomRolePermission()

	testCases := []struct {
		name          string
		roleID        int32
		permissionID  int32
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:         "OK",
			roleID:       rolePermission.RoleID,
			permissionID: rolePermission.PermissionID,
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
					Return([]string{"VIEW_SCREEN_ROLE_PERMISSION"}, nil)
				arg := db.DeleteRolePermissionParams{
					RoleID:       rolePermission.RoleID,
					PermissionID: rolePermission.PermissionID,
				}
				store.EXPECT().
					DeleteRolePermission(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:         "NotFound",
			roleID:       rolePermission.RoleID,
			permissionID: rolePermission.PermissionID,
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
					Return([]string{"VIEW_SCREEN_ROLE_PERMISSION"}, nil)
				arg := db.DeleteRolePermissionParams{
					RoleID:       rolePermission.RoleID,
					PermissionID: rolePermission.PermissionID,
				}
				store.EXPECT().
					DeleteRolePermission(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(sql.ErrNoRows)
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

			url := fmt.Sprintf("/role_permissions/%d/%d", tc.roleID, tc.permissionID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListRolePermissionsAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 10
	rolePermissions := make([]db.RolePermission, n)
	for i := 0; i < n; i++ {
		rolePermissions[i] = randomRolePermission()
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
					Return([]string{"VIEW_SCREEN_ROLE_PERMISSION"}, nil)
				arg := db.ListRolePermissionsParams{
					Limit:  10,
					Offset: 0,
				}
				store.EXPECT().
					ListRolePermissions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(rolePermissions, nil)
				store.EXPECT().
					CountRolePermissions(gomock.Any()).
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

			url := "/role_permissions" + tc.query
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetRolePermissionAPI(t *testing.T) {
	user, _ := randomUser(t)
	rolePermission := randomRolePermission()

	testCases := []struct {
		name          string
		roleID        int32
		permissionID  int32
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:         "OK",
			roleID:       rolePermission.RoleID,
			permissionID: rolePermission.PermissionID,
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
					Return([]string{"VIEW_SCREEN_ROLE_PERMISSION"}, nil)
				arg := db.GetRolePermissionParams{
					RoleID:       rolePermission.RoleID,
					PermissionID: rolePermission.PermissionID,
				}
				store.EXPECT().
					GetRolePermission(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(rolePermission, nil)
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

			url := fmt.Sprintf("/role_permissions/%d/%d", tc.roleID, tc.permissionID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateRolePermissionAPI(t *testing.T) {
	user, _ := randomUser(t)
	rolePermission := randomRolePermission()
	newPermissionID := int32(utils.RandomInt(1, 1000))

	testCases := []struct {
		name          string
		roleID        int32
		permissionID  int32
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:         "OK",
			roleID:       rolePermission.RoleID,
			permissionID: rolePermission.PermissionID,
			body: gin.H{
				"permission_id": newPermissionID,
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
					Return([]string{"VIEW_SCREEN_ROLE_PERMISSION"}, nil)
				arg := db.UpdateRolePermissionParams{
					RoleID:         rolePermission.RoleID,
					PermissionID:   rolePermission.PermissionID,
					PermissionID_2: newPermissionID,
				}
				store.EXPECT().
					UpdateRolePermission(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(rolePermission, nil)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/role_permissions/%d/%d", tc.roleID, tc.permissionID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomRolePermission() db.RolePermission {
	return db.RolePermission{
		RoleID:       int32(utils.RandomInt(1, 1000)),
		PermissionID: int32(utils.RandomInt(1, 1000)),
	}
}
