package handler_test

import (
	"context"
	"os"
	"testing"
)

// TestMain terminates the shared testcontainers Postgres instances created by
// the DB-backed admin student/results test envs. sync.Once keeps each container
// alive across tests in the package; without this they leak after the run.
func TestMain(m *testing.M) {
	code := m.Run()

	ctx := context.Background()
	if adminStuDBEnv != nil {
		if adminStuDBEnv.pool != nil {
			adminStuDBEnv.pool.Close()
		}
		if adminStuDBEnv.rdb != nil {
			adminStuDBEnv.rdb.Close()
		}
		if adminStuDBEnv.mr != nil {
			adminStuDBEnv.mr.Close()
		}
		if adminStuDBEnv.pgContainer != nil {
			_ = adminStuDBEnv.pgContainer.Terminate(ctx)
		}
	}
	if adminResultsDBEnv != nil {
		if adminResultsDBEnv.pool != nil {
			adminResultsDBEnv.pool.Close()
		}
		if adminResultsDBEnv.rdb != nil {
			adminResultsDBEnv.rdb.Close()
		}
		if adminResultsDBEnv.mr != nil {
			adminResultsDBEnv.mr.Close()
		}
		if adminResultsDBEnv.pgContainer != nil {
			_ = adminResultsDBEnv.pgContainer.Terminate(ctx)
		}
	}

	if adminProductDBEnv != nil {
		if adminProductDBEnv.pool != nil {
			adminProductDBEnv.pool.Close()
		}
		if adminProductDBEnv.rdb != nil {
			adminProductDBEnv.rdb.Close()
		}
		if adminProductDBEnv.mr != nil {
			adminProductDBEnv.mr.Close()
		}
		if adminProductDBEnv.pgContainer != nil {
			_ = adminProductDBEnv.pgContainer.Terminate(ctx)
		}
	}

	if regionDBEnv != nil {
		if regionDBEnv.pool != nil {
			regionDBEnv.pool.Close()
		}
		if regionDBEnv.rdb != nil {
			regionDBEnv.rdb.Close()
		}
		if regionDBEnv.mr != nil {
			regionDBEnv.mr.Close()
		}
		if regionDBEnv.pgContainer != nil {
			_ = regionDBEnv.pgContainer.Terminate(ctx)
		}
	}

	os.Exit(code)
}