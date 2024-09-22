package token

import (
	"github.com/bolusarz/urlmini/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	userId := util.RandomInt(1, 10)
	duration := time.Minute

	token, payload, err := maker.CreateToken(userId, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)
	require.Equal(t, userId, payload.UserID)

	parsedPayload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.Equal(t, payload.ID, parsedPayload.ID)
	require.Equal(t, payload.UserID, parsedPayload.UserID)
	require.WithinDuration(t, payload.IssuedAt, parsedPayload.IssuedAt, duration)
	require.WithinDuration(t, payload.ExpireAt, parsedPayload.ExpireAt, duration)

}
