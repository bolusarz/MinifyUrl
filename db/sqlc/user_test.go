package db

import (
	"context"
	"github.com/bolusarz/urlmini/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) (user User, password string) {
	password = util.RandomString(8)

	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = User{
		Username:       util.RandomUsername(),
		FirstName:      util.RandomName(),
		LastName:       util.RandomName(),
		Email:          util.RandomEmail(),
		HashedPassword: hashedPassword,
	}

	return
}

func createRandomDbUser(t *testing.T) User {
	user, _ := createRandomUser(t)
	arg := CreateUserParams{
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
	}
	createdUser, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)

	require.NotEmpty(t, createdUser)
	require.Equal(t, user.HashedPassword, createdUser.HashedPassword)
	requireBodyMatchUser(t, user, createdUser)

	require.NotZero(t, createdUser.CreatedAt)
	require.NotZero(t, createdUser.ID)

	return createdUser
}

func TestQueries_CreateUser(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "OK",
			buildStubs: func(t *testing.T) {
				createRandomDbUser(t)
			},
		},
		{
			name: "EmailExists",
			buildStubs: func(t *testing.T) {
				user, _ := createRandomUser(t)
				arg := CreateUserParams{
					Username:       user.Username,
					FirstName:      user.FirstName,
					LastName:       user.LastName,
					Email:          user.Email,
					HashedPassword: user.HashedPassword,
				}
				_, err := testQueries.CreateUser(context.Background(), arg)
				require.NoError(t, err)

				user2, _ := createRandomUser(t)
				arg = CreateUserParams{
					Username:       user2.Username,
					FirstName:      user2.FirstName,
					LastName:       user2.LastName,
					Email:          user.Email,
					HashedPassword: user2.HashedPassword,
				}

				_, err = testQueries.CreateUser(context.Background(), arg)
				require.Error(t, err)
				require.ErrorContains(t, err, ErrUniqueViolation.Code)
			},
		},
		{
			name: "UsernameExists",
			buildStubs: func(t *testing.T) {
				user, _ := createRandomUser(t)
				arg := CreateUserParams{
					Username:       user.Username,
					FirstName:      user.FirstName,
					LastName:       user.LastName,
					Email:          user.Email,
					HashedPassword: user.HashedPassword,
				}
				_, err := testQueries.CreateUser(context.Background(), arg)
				require.NoError(t, err)

				user2, _ := createRandomUser(t)
				arg = CreateUserParams{
					Username:       user.Username,
					FirstName:      user2.FirstName,
					LastName:       user2.LastName,
					Email:          user2.Email,
					HashedPassword: user2.HashedPassword,
				}

				_, err = testQueries.CreateUser(context.Background(), arg)
				require.Error(t, err)
				require.ErrorContains(t, err, ErrUniqueViolation.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(t)
		})
	}
}

func TestQueries_GetUser(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "By ID: OK",
			buildStubs: func(t *testing.T) {
				user := createRandomDbUser(t)
				fetchedUser, err := testQueries.GetUserById(context.Background(), user.ID)
				require.NoError(t, err)

				require.Equal(t, user.ID, fetchedUser.ID)
				require.Equal(t, user.HashedPassword, fetchedUser.HashedPassword)
				requireBodyMatchUser(t, user, fetchedUser)
				require.WithinDuration(t, user.CreatedAt, fetchedUser.CreatedAt, time.Second)
			},
		},
		{
			name: "By Username: OK",
			buildStubs: func(t *testing.T) {
				user := createRandomDbUser(t)
				fetchedUser, err := testQueries.GetUser(context.Background(), user.Username)
				require.NoError(t, err)

				require.Equal(t, user.ID, fetchedUser.ID)
				require.Equal(t, user.HashedPassword, fetchedUser.HashedPassword)
				requireBodyMatchUser(t, user, fetchedUser)
				require.WithinDuration(t, user.CreatedAt, fetchedUser.CreatedAt, time.Second)
			},
		},
		{
			name: "By Email: OK",
			buildStubs: func(t *testing.T) {
				user := createRandomDbUser(t)
				fetchedUser, err := testQueries.GetUser(context.Background(), user.Email)
				require.NoError(t, err)

				require.Equal(t, user.ID, fetchedUser.ID)
				require.Equal(t, user.HashedPassword, fetchedUser.HashedPassword)
				requireBodyMatchUser(t, user, fetchedUser)
				require.WithinDuration(t, user.CreatedAt, fetchedUser.CreatedAt, time.Second)
			},
		},
		{
			name: "ID: Does Not Exists",
			buildStubs: func(t *testing.T) {
				_, err := testQueries.GetUserById(context.Background(), -1)
				require.Error(t, err)
				require.EqualError(t, err, ErrRecordNotFound.Error())
			},
		},
		{
			name: "Email/Username: Does Not Exists",
			buildStubs: func(t *testing.T) {
				_, err := testQueries.GetUser(context.Background(), "")
				require.Error(t, err)
				require.EqualError(t, err, ErrRecordNotFound.Error())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(t)
		})
	}

}

func TestQueries_UpdateUser(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "OK",
			buildStubs: func(t *testing.T) {
				user := createRandomDbUser(t)
				newUser, _ := createRandomUser(t)
				arg := UpdateUserParams{
					FirstName: newUser.FirstName,
					LastName:  newUser.LastName,
					Email:     newUser.Email,
					Username:  newUser.Username,
					ID:        user.ID,
				}

				updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
				require.NoError(t, err)
				require.NotEmpty(t, updatedUser)

				require.Equal(t, user.ID, updatedUser.ID)
				requireBodyMatchUser(t, newUser, updatedUser)
			},
		},
		{
			name: "UserDoesNotExists",
			buildStubs: func(t *testing.T) {
				user, _ := createRandomUser(t)
				arg := UpdateUserParams{
					FirstName: user.FirstName,
					LastName:  user.LastName,
					Email:     user.Email,
					Username:  user.Username,
					ID:        -1,
				}

				_, err := testQueries.UpdateUser(context.Background(), arg)

				require.Error(t, err)
				require.EqualError(t, err, ErrRecordNotFound.Error())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(t)
		})
	}
}

func requireBodyMatchUser(t *testing.T, user User, otherUser User) {
	require.Equal(t, user.Username, otherUser.Username)
	require.Equal(t, user.FirstName, otherUser.FirstName)
	require.Equal(t, user.LastName, otherUser.LastName)
	require.Equal(t, user.Email, otherUser.Email)
}
