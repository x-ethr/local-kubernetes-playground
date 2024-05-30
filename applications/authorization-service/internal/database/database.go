package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ssl() bool {
	if (os.Getenv("PGSSLMODE") != "disable" && os.Getenv("PGSSLMODE") == "") && os.Getenv("PGSSLROOTCERT") != "" {
		return true
	}

	return false
}

func dsn() (v string) {
	if ssl() {
		v = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s connect_timeout=%s application_name=%s sslmode=%s sslrootcert=%s TimeZone=%s",
			os.Getenv("PGHOST"), os.Getenv("PGPORT"), os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"), os.Getenv("PGCONNECT_TIMEOUT"), os.Getenv("PGAPPNAME"), os.Getenv("PGSSLMODE"), os.Getenv("PGSSLROOTCERT"), os.Getenv("PGTZ"),
		)
	} else {
		v = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s connect_timeout=%s application_name=%s TimeZone=%s",
			os.Getenv("PGHOST"), os.Getenv("PGPORT"), os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"), os.Getenv("PGCONNECT_TIMEOUT"), os.Getenv("PGAPPNAME"), os.Getenv("PGTZ"),
		)
	}

	return
}

var pool atomic.Pointer[pgxpool.Pool]

// Connection establishes a connection to the database using pgxpool.
// If a connection pool does not exist, a new one is created and stored in the pool variable.
// Returns a connection from the connection pool.
// If an error occurs during connection creation, nil and the error are returned.
func Connection(ctx context.Context) (*pgxpool.Conn, error) {
	if pool.Load() == nil {
		configuration, e := pgxpool.ParseConfig(dsn())
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Generate Configuration from DSN String", slog.String("error", e.Error()))
			return nil, e
		}

		instance, e := pgxpool.NewWithConfig(ctx, configuration)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Establish Pool Connection to Database", slog.String("error", e.Error()))
			return nil, e
		}

		pool.Store(instance)
	}

	return pool.Load().Acquire(ctx)
}
