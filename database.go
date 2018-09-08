package main

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql" // source the driver for database/sql but don't call it directly
)

var (
	DBTYPE = "mysql"
)

type DBInstance struct {
	SQLSession *sql.DB
}

/*
	create a new DBInstance struct and return it to caller
	caller is expected to close the database instance dbi.Close()
*/
func NewDBI() (*DBInstance, error) {
	dbi := new(DBInstance)
	err := dbi.ConnectDB()
	if err != nil {
		return nil, err
	}
	return dbi, nil
}

// ConnectDB creates a new database session.  Caller needs to call Close() when done
func (dbi *DBInstance) ConnectDB() error {
	sess, err := sql.Open(DBTYPE, dbURL)
	if err != nil {
		return errors.New("can not connect to database: " + err.Error())
	}
	dbi.SQLSession = sess
	dbi.SQLSession.SetMaxOpenConns(1) // make sure there is only one session open with database at a time
	return nil
}

/*
	close the database session
*/
func (dbi *DBInstance) Close() error {
	if dbi.SQLSession != nil {
		err := dbi.SQLSession.Close()
		if err != nil {
			return errors.New("can not close libpq session: " + err.Error())
		}
	}
	return nil
}

/*####################################################################*/
/*
	QUERY FUNCTIONS
*/

/*
	Return row set from query and expects caller to handle error like sql.ErrNoRows
*/
func (dbi *DBInstance) GetRowSet(qstring string) (*sql.Rows, error) {
	return dbi.SQLSession.Query(qstring)

}

/*
	return string slice from rowset
*/
func (dbi *DBInstance) GetStringList(qstring string) ([]string, error) {
	result := make([]string, 0)
	rows, err := dbi.GetRowSet(qstring)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return result, err
		}
		result = append(result, s)
	}
	return result, nil
}

/*
	return int slice from rowset
*/
func (dbi *DBInstance) GetIntList(qstring string) ([]int, error) {
	result := make([]int, 0)
	rows, err := dbi.GetRowSet(qstring)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		var s int
		err := rows.Scan(&s)
		if err != nil {
			return result, err
		}
		result = append(result, s)
	}
	return result, nil
}

/*
	Assumed single row/column query result of integer type and returns that value
*/
func (dbi *DBInstance) GetIntValue(qstring string) (int, error) {
	var v int
	err := dbi.SQLSession.QueryRow(qstring).Scan(&v)
	if err != nil {
		return v, err
	}
	return v, nil
}

/*
	Assumed single row/column query result of string type and returns that value
*/
func (dbi *DBInstance) GetStringValue(qstring string) (string, error) {
	var v string
	err := dbi.SQLSession.QueryRow(qstring).Scan(&v)
	if err != nil {
		return v, err
	}
	return v, nil
}
