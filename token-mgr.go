package main



import (
	"database/sql"
	"fmt"
	"errors"
)
var(
	TokenTable = "users"
)

type TokenTuple struct {
	UserID int `json:"userid"`
	UserName string `json:"username"`
	RefreshToken string `json:"refreshtoken"`
}

var NoTokenFound = errors.New("Token Manager: No Refresh Token Found")

/*
	Fetch the user refresh token from database
	if no token found then return NoTokenFound
*/
func GetToken(u string) (*TokenTuple, error) {

	t := new(TokenTuple)
	dbi, err := NewDBI()
	if err != nil {
		return t, err
	}
	defer dbi.Close()


	q := fmt.Sprintf("SELECT * FROM %s WHERE username='%s' limit 1", TokenTable, u)
	err = dbi.SQLSession.QueryRow(q).Scan( &t.UserID, &t.UserName, &t.RefreshToken )
	if err == sql.ErrNoRows {
		return t, NoTokenFound
	} else if err != nil {
		fmt.Printf("GetToken() Error when running query %s: %s\n", q, err)
		return t, err
	}
	return t, nil
}

/*
	If user does not exist then insert token
	if user exists then update token
*/
func (t *TokenTuple) UpdateToken() error {
	var q string
	var dberr error

	dbi, err := NewDBI()
	if err != nil {
		return err
	}
	defer dbi.Close()

	tNew, err := GetToken(t.UserName)
	if err == NoTokenFound {
		q = fmt.Sprintf("INSERT INTO %s values(DEFAULT,'%s','%s')", TokenTable, t.UserName, t.RefreshToken)
		_, dberr = dbi.SQLSession.Exec(q)
	}else {
		q = fmt.Sprintf("UPDATE %s SET username='%s', refreshtoken='%s' WHERE id = %d", TokenTable, t.UserName, t.RefreshToken, tNew.UserID)
		_, dberr = dbi.SQLSession.Exec(q)
	}

	if dberr != nil {
		fmt.Printf("UpdateToken() Error when running query %s: %s\n", q, dberr)
		return dberr
	}
	return nil
}
