package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/bolusarz/urlmini/db/mock"
	db "github.com/bolusarz/urlmini/db/sqlc"
	"github.com/bolusarz/urlmini/token"
	"github.com/bolusarz/urlmini/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createRandomLink(userId int64) db.Link {
	link := db.Link{
		UserID:    userId,
		Code:      util.RandomCode(),
		Link:      util.RandomLink(),
		CreatedAt: time.Now(),
		ID:        util.RandomInt(1, 100),
		Active:    pgtype.Bool{Bool: true, Valid: true},
	}

	return link
}

func TestCreateLink(t *testing.T) {
	user, _ := randomUser(t)
	user.ID = util.RandomInt(1, 100)
	link := createRandomLink(user.ID)

	testCases := []struct {
		name          string
		payload       gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			payload: gin.H{
				"link": link.Link,
				"code": link.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				args := db.CreateLinkParams{
					Code:   link.Code,
					Link:   link.Link,
					UserID: link.UserID,
				}
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateLink(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(link, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchLink(t, recorder.Body, link)
			},
		},
		{
			name: "NoCodeProvided",
			payload: gin.H{
				"link": link.Link,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateLink(gomock.Any(), gomock.Any()).
					Times(1).
					Return(link, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "NoLinkProvided",
			payload: gin.H{
				"link": "",
				"code": link.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateLink(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			payload: gin.H{
				"link": link.Link,
				"code": link.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateLink(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Link{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			jsonBody, err := json.Marshal(tc.payload)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/links", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetLinkByCode(t *testing.T) {
	user, _ := randomUser(t)
	user.ID = util.RandomInt(1, 20)
	link := createRandomLink(user.ID)
	inActiveLink := createRandomLink(user.ID)
	inActiveLink.Active = pgtype.Bool{Bool: false, Valid: true}

	testCases := []struct {
		name          string
		payload       gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			payload: gin.H{
				"code": link.Code,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLinkByCode(gomock.Any(), gomock.Eq(link.Code)).
					Times(1).
					Return(link, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusPermanentRedirect)
			},
		},
		{
			name: "CodeIsNotActive",
			payload: gin.H{
				"code": inActiveLink.Code,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLinkByCode(gomock.Any(), gomock.Eq(inActiveLink.Code)).
					Times(1).
					Return(inActiveLink, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusNotFound)
			},
		},
		{
			name: "CodeDoesNotExist",
			payload: gin.H{
				"code": "notexists",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLinkByCode(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Link{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusNotFound)
			},
		},
		{
			name: "BadRequest",
			payload: gin.H{
				"code": "1234ha",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLinkByCode(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			payload: gin.H{
				"code": link.Code,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetLinkByCode(gomock.Any(), gomock.Eq(link.Code)).
					Times(1).
					Return(db.Link{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusInternalServerError)
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

			request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%v", tc.payload["code"]), nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)

		})
	}

}

func TestGetLinkById(t *testing.T) {
	user, _ := randomUser(t)
	user.ID = util.RandomInt(1, 20)

	link := createRandomLink(user.ID)

	testCases := []struct {
		name          string
		payload       gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			payload: gin.H{
				"id": link.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(link, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusOK)
				requireBodyMatchLink(t, recorder.Body, link)
			},
		},
		{
			name: "NotFound",
			payload: gin.H{
				"id": util.RandomInt(1, 1000000),
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Link{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusNotFound)
			},
		},
		{
			name: "AccessForbidden",
			payload: gin.H{
				"id": link.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID+1, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID+1)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(link, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusForbidden)
			},
		},
		{
			name: "BadRequest",
			payload: gin.H{
				"id": "notanumber",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "InternalServerError",
			payload: gin.H{
				"id": link.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(db.Link{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusInternalServerError)
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

			request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/links/%v", tc.payload["id"]), nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestChangeCode(t *testing.T) {
	user, _ := randomUser(t)
	user.ID = util.RandomInt(1, 20)

	link := createRandomLink(user.ID)
	updatedLink := db.Link{
		ID:        link.ID,
		Link:      link.Link,
		UserID:    link.UserID,
		Code:      util.RandomCode(),
		Active:    link.Active,
		CreatedAt: link.CreatedAt,
	}
	testCases := []struct {
		name          string
		payload       gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			payload: gin.H{
				"id":   link.ID,
				"code": updatedLink.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(link, nil)

				args := db.UpdateCodeParams{
					ID:   link.ID,
					Code: updatedLink.Code,
				}

				store.EXPECT().
					UpdateCode(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(updatedLink, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusOK)
				requireBodyMatchLink(t, recorder.Body, updatedLink)
			},
		},
		{
			name: "UserDoesNotOwnLink",
			payload: gin.H{
				"id":   link.ID,
				"code": updatedLink.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID+1, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID+1)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(link, nil)

				store.EXPECT().
					UpdateCode(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusForbidden)
			},
		},
		{
			name: "LinkDoesNotExist",
			payload: gin.H{
				"id":   link.ID,
				"code": updatedLink.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(db.Link{}, db.ErrRecordNotFound)

				store.EXPECT().
					UpdateCode(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusNotFound)
			},
		},
		{
			name: "InvalidID",
			payload: gin.H{
				"id":   "notANumber",
				"code": updatedLink.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(0)

				store.EXPECT().
					UpdateCode(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "InvalidCode",
			payload: gin.H{
				"id":   link.ID,
				"code": "&**yuh12",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(0)

				store.EXPECT().
					UpdateCode(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "CouldNotFetchLink",
			payload: gin.H{
				"id":   link.ID,
				"code": updatedLink.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(db.Link{}, sql.ErrConnDone)

				store.EXPECT().
					UpdateCode(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusInternalServerError)
			},
		},
		{
			name: "CouldNotUpdateLink",
			payload: gin.H{
				"id":   link.ID,
				"code": updatedLink.Code,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(link, nil)

				args := db.UpdateCodeParams{
					ID:   link.ID,
					Code: updatedLink.Code,
				}

				store.EXPECT().
					UpdateCode(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(db.Link{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusInternalServerError)
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

			jsonBody, err := json.Marshal(tc.payload)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/links/%v", tc.payload["id"]),
				bytes.NewBuffer(jsonBody),
			)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestToggleLinkStatus(t *testing.T) {
	user, _ := randomUser(t)
	user.ID = util.RandomInt(1, 20)

	link := createRandomLink(user.ID)
	updatedLink := db.Link{
		ID:     link.ID,
		Link:   link.Link,
		UserID: link.UserID,
		Code:   link.Code,
		Active: pgtype.Bool{
			Bool:  !link.Active.Bool,
			Valid: true,
		},
		CreatedAt: link.CreatedAt,
	}
	testCases := []struct {
		name          string
		payload       gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			payload: gin.H{
				"id": link.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(link, nil)

				args := db.ToggleStatusParams{
					ID: link.ID,
				}

				store.EXPECT().
					ToggleStatus(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(updatedLink, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusOK)
				requireBodyMatchLink(t, recorder.Body, updatedLink)
			},
		},
		{
			name: "UserDoesNotOwnLink",
			payload: gin.H{
				"id": link.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID+1, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID+1)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(link, nil)

				store.EXPECT().
					ToggleStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusForbidden)
			},
		},
		{
			name: "LinkDoesNotExist",
			payload: gin.H{
				"id": link.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(db.Link{}, db.ErrRecordNotFound)

				store.EXPECT().
					ToggleStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusNotFound)
			},
		},
		{
			name: "InvalidID",
			payload: gin.H{
				"id": "notANumber",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(0)

				store.EXPECT().
					ToggleStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
		{
			name: "CouldNotFetchLink",
			payload: gin.H{
				"id": link.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(db.Link{}, sql.ErrConnDone)

				store.EXPECT().
					ToggleStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusInternalServerError)
			},
		},
		{
			name: "CouldNotUpdateLink",
			payload: gin.H{
				"id": link.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinkById(gomock.Any(), gomock.Eq(link.ID)).
					Times(1).
					Return(link, nil)

				args := db.ToggleStatusParams{
					ID: link.ID,
				}

				store.EXPECT().
					ToggleStatus(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(db.Link{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusInternalServerError)
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

			request, err := http.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/links/%v/toggle", tc.payload["id"]),
				nil,
			)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetLinks(t *testing.T) {
	user, _ := randomUser(t)
	user.ID = util.RandomInt(1, 20)

	n := 10
	links := make([]db.Link, n)

	for i := 0; i < n; i++ {
		links[i] = createRandomLink(user.ID)
	}

	type Query struct {
		PageID   int32
		PageSize int32
	}

	testCases := []struct {
		name          string
		query         *Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: &Query{
				PageID:   1,
				PageSize: 10,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				args := db.GetLinksByUserParams{
					UserID: user.ID,
					Offset: 0,
					Limit:  int32(n),
				}

				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinksByUser(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(links, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusOK)
				requireBodyMatchLinks(t, recorder.Body, links)
			},
		},
		{
			name:  "NoQueryPassed",
			query: nil,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				args := db.GetLinksByUserParams{
					UserID: user.ID,
					Offset: 0,
					Limit:  10,
				}

				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinksByUser(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(links, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusOK)
				requireBodyMatchLinks(t, recorder.Body, links)
			},
		},
		{
			name:  "InternalServerError",
			query: nil,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserById(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetLinksByUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusInternalServerError)
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

			request, err := http.NewRequest(http.MethodGet, "/links", nil)
			require.NoError(t, err)

			if tc.query != nil {
				q := request.URL.Query()
				q.Add("page_id", fmt.Sprintf("%d", tc.query.PageID))
				q.Add("page_size", fmt.Sprintf("%d", tc.query.PageSize))
				request.URL.RawQuery = q.Encode()
			}

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchLink(t *testing.T, body *bytes.Buffer, link db.Link) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	gotLink := response["data"].(map[string]interface{})

	compareMapToLink(t, gotLink, link)
}

func requireBodyMatchLinks(t *testing.T, body *bytes.Buffer, links []db.Link) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	if gotLinks, ok := response["data"].([]map[string]any); ok {
		for i, gotLink := range gotLinks {
			compareMapToLink(t, gotLink, links[i])
		}
	}

}

func compareMapToLink(t *testing.T, payload map[string]any, link db.Link) {
	require.Equal(t, link.Link, payload["link"])
	require.Equal(t, link.Code, payload["code"])
	require.Equal(t, link.UserID, int64(payload["user_id"].(float64)))
	require.NotZero(t, payload["created_at"])
}
