package db

import (
	"context"
	"github.com/bolusarz/urlmini/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"os"
	"testing"
)

var testQueries *Queries

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")

	if err != nil {
		panic(err)
	}

	testDB, err := pgxpool.New(context.Background(), config.DBSource)

	if err != nil {
		panic(errors.Errorf("Error connecting to database: %v", err))
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
