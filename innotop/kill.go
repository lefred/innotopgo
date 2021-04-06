package innotop

import (
	"database/sql"
	"errors"

	"github.com/lefred/innotopgo/db"
)

func KillQuery(mydb *sql.DB, thread_id string) error {
	// TODO it works only with conn_id
	stmt := `select pps.PROCESSLIST_ID AS conn_id from performance_schema.threads pps where thread_id = ` + thread_id + ` LIMIT 1`
	rows, err := db.Query(mydb, stmt)

	if err != nil {
		return (err)
	}
	_, data, _ := db.GetData(rows)
	var conn_id string
	if len(data) < 1 {
		err = errors.New("not found")
		return err
	}
	for _, row := range data {
		conn_id = row[0]
	}
	stmt = `kill query ` + conn_id
	err = db.RunQuery(mydb, stmt)
	return err
}
