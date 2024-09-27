package db

import (
	"context"
	"github.com/bolusarz/urlmini/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomDbSession(t *testing.T) Session {
	user := createRandomDbUser(t)
	token, tokenPayload, err := tokenMaker.CreateToken(user.ID, time.Minute)
	require.NoError(t, err)

	arg := CreateSessionParams{
		ID:           tokenPayload.ID,
		UserID:       tokenPayload.UserID,
		RefreshToken: token,
		ClientIp:     util.RandomIP(),
		UserAgent:    util.RandomString(10),
		IsBlocked:    false,
		ExpiresAt:    tokenPayload.ExpireAt,
	}

	session, err := testQueries.CreateSession(context.Background(), arg)

	require.NoError(t, err)

	require.Equal(t, session.ID, tokenPayload.ID)
	require.NotZero(t, session.CreatedAt)

	require.Equal(t, arg.UserID, session.UserID)
	require.Equal(t, arg.RefreshToken, session.RefreshToken)
	require.Equal(t, arg.ClientIp, session.ClientIp)
	require.Equal(t, arg.UserAgent, session.UserAgent)
	require.Equal(t, arg.IsBlocked, session.IsBlocked)
	require.WithinDuration(t, arg.ExpiresAt, session.ExpiresAt, time.Minute)

	return session
}

func TestQueries_CreateSession(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "OK",
			buildStubs: func(t *testing.T) {
				createRandomDbSession(t)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.buildStubs)
	}
}

func TestQueries_BlockSession(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "OK",
			buildStubs: func(t *testing.T) {
				session := createRandomDbSession(t)

				blockedSession, err := testQueries.BlockSession(context.Background(), session.ID)
				require.NoError(t, err)

				require.Equal(t, session.ID, blockedSession.ID)
				require.Equal(t, session.RefreshToken, blockedSession.RefreshToken)
				require.Equal(t, session.UserID, blockedSession.UserID)
				require.Equal(t, session.ClientIp, blockedSession.ClientIp)
				require.Equal(t, session.UserAgent, blockedSession.UserAgent)
				require.Equal(t, session.CreatedAt, blockedSession.CreatedAt)
				require.Equal(t, session.ExpiresAt, blockedSession.ExpiresAt)
				require.True(t, blockedSession.IsBlocked)

			},
		},
		{
			name: "DoesNotExist",
			buildStubs: func(t *testing.T) {
				blockedSession, err := testQueries.BlockSession(context.Background(), uuid.New())
				require.Error(t, err)
				require.EqualError(t, err, ErrRecordNotFound.Error())
				require.Empty(t, blockedSession)

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.buildStubs)
	}
}

func TestQueries_GetSessions(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "ActiveSessions:OK",
			buildStubs: func(t *testing.T) {
				user := createRandomDbUser(t)
				runs := 5
				n := 3
				sessions := make([]Session, n)
				index := n - 1

				for i := 0; i < runs; i++ {
					token, tokenPayload, err := tokenMaker.CreateToken(user.ID, time.Minute)
					require.NoError(t, err)

					arg := CreateSessionParams{
						ID:           tokenPayload.ID,
						UserID:       tokenPayload.UserID,
						RefreshToken: token,
						ClientIp:     util.RandomIP(),
						UserAgent:    util.RandomString(10),
						IsBlocked:    false,
						ExpiresAt:    tokenPayload.ExpireAt,
					}

					session, err := testQueries.CreateSession(context.Background(), arg)
					require.NoError(t, err)

					if i%2 == 0 {
						sessions[index] = session
						index -= 1
					} else {
						_, err := testQueries.BlockSession(context.Background(), session.ID)
						require.NoError(t, err)
					}
				}

				args := GetActiveSessionsParams{
					UserID: user.ID,
					Limit:  int32(n),
					Offset: 0,
				}
				fetchedSessions, err := testQueries.GetActiveSessions(context.Background(), args)
				require.NoError(t, err)

				require.Equal(t, fetchedSessions, sessions)
			},
		},
		{
			name: "Sessions:OK",
			buildStubs: func(t *testing.T) {
				user := createRandomDbUser(t)
				n := 3
				sessions := make([]Session, n)

				for i := 0; i < n; i++ {
					token, tokenPayload, err := tokenMaker.CreateToken(user.ID, time.Minute)
					require.NoError(t, err)

					arg := CreateSessionParams{
						ID:           tokenPayload.ID,
						UserID:       tokenPayload.UserID,
						RefreshToken: token,
						ClientIp:     util.RandomIP(),
						UserAgent:    util.RandomString(10),
						IsBlocked:    false,
						ExpiresAt:    tokenPayload.ExpireAt,
					}

					session, err := testQueries.CreateSession(context.Background(), arg)
					require.NoError(t, err)

					sessions[n-i-1] = session
				}

				args := GetSessionsParams{
					UserID: user.ID,
					Limit:  int32(n),
					Offset: 0,
				}
				fetchedSessions, err := testQueries.GetSessions(context.Background(), args)
				require.NoError(t, err)

				require.Equal(t, fetchedSessions, sessions)
			},
		},
		{
			name: "DoesNotExist",
			buildStubs: func(t *testing.T) {
				args := GetActiveSessionsParams{
					UserID: -1,
					Limit:  10,
					Offset: 0,
				}
				fetchedSessions, err := testQueries.GetActiveSessions(context.Background(), args)
				require.NoError(t, err)
				require.Len(t, fetchedSessions, 0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.buildStubs)
	}
}

func TestQueries_GetSession(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "OK",
			buildStubs: func(t *testing.T) {
				session := createRandomDbSession(t)

				fetchedSession, err := testQueries.GetSession(context.Background(), session.ID)
				require.NoError(t, err)

				require.Equal(t, fetchedSession, session)
			},
		},
		{
			name: "DoesNotExist",
			buildStubs: func(t *testing.T) {
				fetchedSession, err := testQueries.GetSession(context.Background(), uuid.New())
				require.Error(t, err)
				require.EqualError(t, err, ErrRecordNotFound.Error())

				require.Empty(t, fetchedSession)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.buildStubs)
	}
}
