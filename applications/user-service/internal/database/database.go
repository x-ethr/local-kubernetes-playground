package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
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

func Connection(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, dsn())
}
