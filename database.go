package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // source the driver for database/sql but don't call it directly
)

// DBInstance represents the existing database connection
type DBInstance struct {
	SQLSession *sql.DB
	DBURL      string
	DBType     string
}

// NewDBI creates a new databse connection
func NewDBI(dbURL string) (*DBInstance, error) {
	dbi := &DBInstance{nil, dbURL, "mysql"}
	err := dbi.ConnectDB()
	if err != nil {
		return nil, err
	}
	return dbi, nil
}

// ConnectDB creates a new database session.  Caller needs to call Close() when done
func (dbi *DBInstance) ConnectDB() error {
	sess, err := sql.Open(dbi.DBType, dbi.DBURL)
	if err != nil {
		return errors.New("can not connect to database: " + err.Error())
	}
	dbi.SQLSession = sess
	dbi.SQLSession.SetMaxOpenConns(1) // make sure there is only one session open with database at a time
	return nil
}

// Close the database session
func (dbi *DBInstance) Close() error {
	if dbi.SQLSession != nil {
		err := dbi.SQLSession.Close()
		if err != nil {
			return errors.New("can not close libpq session: " + err.Error())
		}
	}
	return nil
}

// CreateSchema will create database tables and import words if tables do not exist
func (dbi *DBInstance) CreateSchema() error {

	// check if peers table exists
	v, err := dbi.GetIntValue(SELECT_WORDS_TABLE)
	if err == nil && v > 0 {
		fmt.Println("Words table detected skipping schema installation")
		return nil
	} else {
		fmt.Println("Words table not found starting schema installation")
	}

	// create tables
	_, err = dbi.SQLSession.Exec(CREATE_WORDS_TABLE)
	if err != nil {
		return err
	}
	_, err = dbi.SQLSession.Exec(CREATE_PEERS_TABLE)
	if err != nil {
		return err
	}
	_, err = dbi.SQLSession.Exec(CREATE_SUBJECTS_TABLE)
	if err != nil {
		return err
	}
	_, err = dbi.SQLSession.Exec(CREATE_USERS_TABLE)
	if err != nil {
		return err
	}
	_, err = dbi.SQLSession.Exec(CREAT_SESSIONS_TABLE)
	if err != nil {
		return err
	}
	// insert words
	_, err = dbi.SQLSession.Exec(INSERT_WORDS)
	if err != nil {
		return err
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

// ExecTXQuery creates a transaction and executes given query with args
func (dbi *DBInstance) ExecTXQuery(q string, args ...interface{}) error {
	tx, err := dbi.SQLSession.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}
	return tx.Commit()
}
