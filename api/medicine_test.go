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

func TestCreateMedicineAPI(t *testing.T) {
	user, _ := randomUser(t)
	medicine := randomMedicine()

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
				"name":        medicine.Name,
				"unit":        medicine.Unit,
				"price":       100.0, // Use float for request
				"stock":       medicine.Stock,
				"description": medicine.Description.String,
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
					Return([]string{"VIEW_SCREEN_MEDICINE"}, nil)

				store.EXPECT().
					CreateMedicine(gomock.Any(), gomock.Any()).
					Times(1).
					Return(medicine, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchMedicine(t, recorder.Body, medicine)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"name":  medicine.Name,
				"unit":  medicine.Unit,
				"price": 100.0,
				"stock": medicine.Stock,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateMedicine(gomock.Any(), gomock.Any()).Times(0)
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

			url := "/medicines"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetMedicineAPI(t *testing.T) {
	user, _ := randomUser(t)
	medicine := randomMedicine()

	testCases := []struct {
		name          string
		medicineID    int32
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			medicineID: medicine.ID,
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
					Return([]string{"VIEW_SCREEN_MEDICINE"}, nil)
				store.EXPECT().
					GetMedicine(gomock.Any(), gomock.Eq(medicine.ID)).
					Times(1).
					Return(medicine, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchMedicine(t, recorder.Body, medicine)
			},
		},
		{
			name:       "NotFound",
			medicineID: medicine.ID,
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
					Return([]string{"VIEW_SCREEN_MEDICINE"}, nil)
				store.EXPECT().
					GetMedicine(gomock.Any(), gomock.Eq(medicine.ID)).
					Times(1).
					Return(db.Medicine{}, sql.ErrNoRows)
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

			url := fmt.Sprintf("/medicines/%d", tc.medicineID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomMedicine() db.Medicine {
	return db.Medicine{
		ID:    int32(utils.RandomInt(1, 1000)),
		Name:  utils.RandomString(6),
		Unit:  "tablet", // Use valid enum value
		Price: "100.00",
		Stock: int32(utils.RandomInt(1, 100)),
		Description: sql.NullString{
			String: utils.RandomString(20),
			Valid:  true,
		},
	}
}

func requireBodyMatchMedicine(t *testing.T, body *bytes.Buffer, medicine db.Medicine) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response struct {
		Data db.Medicine `json:"data"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)
	require.Equal(t, medicine, response.Data)
}
