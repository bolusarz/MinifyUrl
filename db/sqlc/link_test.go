package db

import (
	"context"
	"github.com/bolusarz/urlmini/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"testing"
)

func createRandomDbLink(t *testing.T) Link {
	user := createRandomDbUser(t)

	arg := CreateLinkParams{
		Code:   util.RandomCode(),
		Link:   util.RandomLink(),
		UserID: user.ID,
	}

	link, err := testQueries.CreateLink(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, link)

	require.Equal(t, arg.Code, link.Code)
	require.Equal(t, arg.Link, link.Link)
	require.Equal(t, arg.UserID, link.UserID)
	return link
}

func TestQueries_CreateLink(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "OK",
			buildStubs: func(t *testing.T) {
				createRandomDbLink(t)
			},
		},
		{
			name: "CodeExists",
			buildStubs: func(t *testing.T) {
				link := createRandomDbLink(t)

				arg := CreateLinkParams{
					Code:   link.Code,
					Link:   util.RandomLink(),
					UserID: link.UserID,
				}

				_, err := testQueries.CreateLink(context.Background(), arg)
				require.Error(t, err)
				require.ErrorContains(t, err, ErrUniqueViolation.Code)
			},
		},
		{
			name: "SameLinkMultipleCodes",
			buildStubs: func(t *testing.T) {
				link := createRandomDbLink(t)

				arg := CreateLinkParams{
					Code:   util.RandomCode(),
					Link:   link.Link,
					UserID: link.UserID,
				}

				link2, err := testQueries.CreateLink(context.Background(), arg)
				require.NoError(t, err)
				require.NotEmpty(t, link2)

				require.Equal(t, link.Link, link2.Link)

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

func TestQueries_GetLink(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "By ID: OK",
			buildStubs: func(t *testing.T) {
				link := createRandomDbLink(t)

				fetchedLink, err := testQueries.GetLinkById(context.Background(), link.ID)
				require.NoError(t, err)
				require.Equal(t, link, fetchedLink)
			},
		},
		{
			name: "By ID: DoesNotExist",
			buildStubs: func(t *testing.T) {
				_, err := testQueries.GetLinkById(context.Background(), -1)
				require.Error(t, err)
				require.EqualError(t, err, ErrRecordNotFound.Error())
			},
		},
		{
			name: "By Code: OK",
			buildStubs: func(t *testing.T) {
				link := createRandomDbLink(t)

				fetchedLink, err := testQueries.GetLinkByCode(context.Background(), link.Code)
				require.NoError(t, err)
				require.Equal(t, link, fetchedLink)
			},
		},
		{
			name: "By Code: DoesNotExist",
			buildStubs: func(t *testing.T) {
				_, err := testQueries.GetLinkByCode(context.Background(), "")
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

func TestQueries_GetLinks(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "OK",
			buildStubs: func(t *testing.T) {
				n := 5

				for i := 0; i < n; i++ {
					arg := CreateLinkParams{
						Code:   util.RandomCode(),
						Link:   util.RandomLink(),
						UserID: createRandomDbUser(t).ID,
					}
					_, err := testQueries.CreateLink(context.Background(), arg)
					require.NoError(t, err)
				}

				arg := GetLinksParams{
					Limit:  5,
					Offset: 0,
				}
				links, err := testQueries.GetLinks(context.Background(), arg)

				require.NoError(t, err)
				require.NotEmpty(t, links)
				require.Len(t, links, n)
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

func TestQueries_GetLinksByUser(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "OK",
			buildStubs: func(t *testing.T) {
				user := createRandomDbUser(t)

				n := 5

				var link Link
				var err error

				for i := 0; i < n; i++ {
					arg := CreateLinkParams{
						Code:   util.RandomCode(),
						Link:   util.RandomLink(),
						UserID: user.ID,
					}
					link, err = testQueries.CreateLink(context.Background(), arg)
					require.NoError(t, err)
				}

				arg := GetLinksByUserParams{
					UserID: user.ID,
					Limit:  5,
					Offset: 0,
				}
				links, err := testQueries.GetLinksByUser(context.Background(), arg)

				require.NoError(t, err)
				require.NotEmpty(t, links)
				require.Len(t, links, n)
				require.Equal(t, link.Link, links[0].Link)
			},
		},
		{
			name: "DoesNotExist",
			buildStubs: func(t *testing.T) {
				arg := GetLinksByUserParams{
					UserID: -1,
					Limit:  5,
					Offset: 0,
				}

				links, _ := testQueries.GetLinksByUser(context.Background(), arg)
				require.Empty(t, links)
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

func TestQueries_UpdateLink(t *testing.T) {
	testCases := []struct {
		name       string
		buildStubs func(t *testing.T)
	}{
		{
			name: "Update: OK",
			buildStubs: func(t *testing.T) {
				link := createRandomDbLink(t)

				arg := UpdateCodeParams{
					ID:   link.ID,
					Code: util.RandomCode(),
				}
				updateLink, err := testQueries.UpdateCode(context.Background(), arg)
				require.NoError(t, err)

				require.Equal(t, link.Active.Bool, updateLink.Active.Bool)
				require.Equal(t, link.Link, updateLink.Link)
				require.Equal(t, arg.Code, updateLink.Code)
				require.Equal(t, link.ID, updateLink.ID)
			},
		},
		{
			name: "Update: DoesNoExist",
			buildStubs: func(t *testing.T) {
				arg := UpdateCodeParams{
					ID:   -1,
					Code: util.RandomCode(),
				}
				_, err := testQueries.UpdateCode(context.Background(), arg)

				require.Error(t, err)
				require.EqualError(t, err, ErrRecordNotFound.Error())
			},
		},
		{
			name: "Toggle: OK",
			buildStubs: func(t *testing.T) {
				link := createRandomDbLink(t)

				arg := ToggleStatusParams{
					ID: link.ID,
					Active: pgtype.Bool{
						Bool:  !link.Active.Bool,
						Valid: link.Active.Valid,
					},
				}
				updateLink, err := testQueries.ToggleStatus(context.Background(), arg)
				require.NoError(t, err)

				require.Equal(t, !link.Active.Bool, updateLink.Active.Bool)
				require.Equal(t, link.Link, updateLink.Link)
				require.Equal(t, link.Code, updateLink.Code)
				require.Equal(t, link.ID, updateLink.ID)
			},
		},
		{
			name: "Toggle: DoesNoExist",
			buildStubs: func(t *testing.T) {
				link := createRandomDbLink(t)

				arg := ToggleStatusParams{
					ID: -1,
					Active: pgtype.Bool{
						Bool:  !link.Active.Bool,
						Valid: link.Active.Valid,
					},
				}
				_, err := testQueries.ToggleStatus(context.Background(), arg)
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
