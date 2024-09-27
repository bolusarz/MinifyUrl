package api

import (
	"bytes"
	"encoding/json"
	mockdb "github.com/bolusarz/urlmini/db/mock"
	db "github.com/bolusarz/urlmini/db/sqlc"
	"github.com/bolusarz/urlmini/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRenewAccessToken(t *testing.T) {

	//refreshToken, payload, err :=

	testCases := []struct {
		name          string
		payload       func(tokenMaker token.Maker) gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "BadRequest",
			payload: func(tokenMaker token.Maker) gin.H {
				return gin.H{}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "InvalidToken",
			payload: func(tokenMaker token.Maker) gin.H {
				return gin.H{
					"refresh_token": "invalid_token",
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusUnauthorized)
			},
		},
		{
			name: "NoSession",
			payload: func(tokenMaker token.Maker) gin.H {
				refreshToken, _, _ := tokenMaker.CreateToken(1, time.Minute)
				return gin.H{
					"refresh_token": refreshToken,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "SessionExpired",
			payload: func(tokenMaker token.Maker) gin.H {
				refreshToken, _, _ := tokenMaker.CreateToken(1, time.Minute)
				return gin.H{
					"refresh_token": refreshToken,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{
						ExpiresAt: time.Now().Add(-time.Minute),
					}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "SessionBlocked",
			payload: func(tokenMaker token.Maker) gin.H {
				refreshToken, _, _ := tokenMaker.CreateToken(1, time.Minute)
				return gin.H{
					"refresh_token": refreshToken,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{
						ExpiresAt: time.Now().Add(-time.Minute),
						IsBlocked: true,
					}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "OK",
			payload: func(tokenMaker token.Maker) gin.H {
				refreshToken, _, _ := tokenMaker.CreateToken(1, time.Minute)
				return gin.H{
					"refresh_token": refreshToken,
				}
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{
						ExpiresAt: time.Now().Add(time.Minute),
						IsBlocked: false,
					}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBody(t, recorder.Body)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			jsonBody, err := json.Marshal(tc.payload(server.tokenMaker))
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/token/refresh", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBody(t *testing.T, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	responseData := response["data"].(map[string]interface{})

	require.NotEmpty(t, responseData["access_token"])
	require.NotEmpty(t, responseData["access_token_expires_at"])
}
