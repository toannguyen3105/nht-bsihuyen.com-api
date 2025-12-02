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
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	mockdb "github.com/toannguyen3105/nht-bsihuyen.com-api/db/mock"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
	"github.com/toannguyen3105/nht-bsihuyen.com-api/token"
)

func TestAddUserRoleAPI(t *testing.T) {
	user, _ := randomUser(t)
	role := randomRole()
	user.ID = 10
	role.ID = 11

	userRole := db.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}

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
				"user_id": userRole.UserID,
				"role_id": userRole.RoleID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				arg := db.AddRoleForUserParams{
					UserID: userRole.UserID,
					RoleID: userRole.RoleID,
				}
				store.EXPECT().
					AddRoleForUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userRole, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUserRole(t, recorder.Body, userRole)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id": userRole.UserID,
				"role_id": userRole.RoleID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					AddRoleForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserRole{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUserRole",
			body: gin.H{
				"user_id": userRole.UserID,
				"role_id": userRole.RoleID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					AddRoleForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserRole{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AddRoleForUser(gomock.Any(), gomock.Any()).
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

			url := "/user-roles"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetUserRolesAPI(t *testing.T) {
	user, _ := randomUser(t)
	role1 := randomRole()
	role2 := randomRole()
	user.ID = 10
	role1.ID = 11
	role2.ID = 12

	roles := []db.Role{role1, role2}

	testCases := []struct {
		name          string
		userID        int32
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			userID: user.ID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					GetRolesForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(roles, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchRoles(t, recorder.Body, roles)
			},
		},
		{
			name:   "NotFound",
			userID: user.ID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					GetRolesForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.Role{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			userID: user.ID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					GetRolesForUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return([]db.Role{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidUserID",
			userID: 0,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					GetRolesForUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
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

			url := fmt.Sprintf("/users/%d/roles", tc.userID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchUserRole(t *testing.T, body *bytes.Buffer, userRole db.UserRole) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response struct {
		Data db.UserRole `json:"data"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)
	require.Equal(t, userRole, response.Data)
}

func requireBodyMatchRoles(t *testing.T, body *bytes.Buffer, roles []db.Role) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response struct {
		Data []roleResponse `json:"data"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	// Convert expected roles to roleResponse for comparison
	expected := make([]roleResponse, len(roles))
	for i, role := range roles {
		expected[i] = newRoleResponse(role)
	}

	require.Equal(t, expected, response.Data)
}

func TestDeleteUserRoleAPI(t *testing.T) {
	user, _ := randomUser(t)
	role := randomRole()
	user.ID = 10
	role.ID = 11

	userRole := db.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}

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
				"user_id": userRole.UserID,
				"role_id": userRole.RoleID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				arg := db.RemoveRoleForUserParams{
					UserID: userRole.UserID,
					RoleID: userRole.RoleID,
				}
				store.EXPECT().
					RemoveRoleForUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"user_id": userRole.UserID,
				"role_id": userRole.RoleID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					RemoveRoleForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id": userRole.UserID,
				"role_id": userRole.RoleID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					RemoveRoleForUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := "/user-roles"
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateUserRoleAPI(t *testing.T) {
	user, _ := randomUser(t)
	oldRole := randomRole()
	newRole := randomRole()
	user.ID = 10
	oldRole.ID = 11
	newRole.ID = 12

	userRole := db.UserRole{
		UserID: user.ID,
		RoleID: newRole.ID,
	}

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
				"user_id":     user.ID,
				"old_role_id": oldRole.ID,
				"new_role_id": newRole.ID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)

				arg := db.UpdateUserRoleTxParams{
					UserID:    user.ID,
					OldRoleID: oldRole.ID,
					NewRoleID: newRole.ID,
				}
				store.EXPECT().
					UpdateUserRoleTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UpdateUserRoleTxResult{UserRole: userRole}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUserRole(t, recorder.Body, userRole)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id":     user.ID,
				"old_role_id": oldRole.ID,
				"new_role_id": newRole.ID,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					UpdateUserRoleTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UpdateUserRoleTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := "/user-roles"
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListUserRolesAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 10
	userRoles := make([]db.UserRole, n)
	for i := 0; i < n; i++ {
		userRoles[i] = db.UserRole{
			UserID: user.ID,
			RoleID: int32(i + 1),
		}
	}

	type query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: query{
				pageID:   1,
				pageSize: n,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				arg := db.ListUserRolesParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListUserRoles(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userRoles, nil)

				store.EXPECT().
					CountUserRoles(gomock.Any()).
					Times(1).
					Return(int64(n), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUserRoles(t, recorder.Body, userRoles)
			},
		},
		{
			name: "InternalError",
			query: query{
				pageID:   1,
				pageSize: n,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					ListUserRoles(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.UserRole{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageID",
			query: query{
				pageID:   0,
				pageSize: n,
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
					Return([]string{"VIEW_SCREEN_USER_ROLE"}, nil)
				store.EXPECT().
					ListUserRoles(gomock.Any(), gomock.Any()).
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

			url := "/user-roles"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchUserRoles(t *testing.T, body *bytes.Buffer, userRoles []db.UserRole) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	type userRolesResponse struct {
		Meta struct {
			Page       int32 `json:"page"`
			TotalPages int32 `json:"total_pages"`
			TotalCount int64 `json:"total_count"`
		} `json:"meta"`
		Data []db.UserRole `json:"data"`
	}

	var response struct {
		Data userRolesResponse `json:"data"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)
	require.Equal(t, userRoles, response.Data.Data)
}
