package main

import (
	"database/sql"
	"fmt"
)

// UserAccount encapsulate the a token tuple so we can use some the of logic for internal auth
type UserAccount struct {
	Token *TokenTuple
}

// NewUserMustNotExist creates a new user if one does not exist already
func NewUserMustNotExist(user, pass string) (UserAccount, error) {
	u := UserAccount{&TokenTuple{0, user, EncryptedText{pass}}}
	err := u.setUserID()
	if err != nil && err != sql.ErrNoRows {
		return u, fmt.Errorf("database error while creating new user: %s", err)
	}
	if err != sql.ErrNoRows {
		return u, fmt.Errorf("user already exists")
	}

	err = u.Token.UpdateToken() // add the new user to the database
	if err != nil {
		return u, err
	}
	return u, nil
}

// GetExistingUser returns existing user from database
func GetExistingUser(user string) (UserAccount, error) {
	u := UserAccount{}
	var err error
	u.Token, err = GetToken(user)
	return u, err
}

func (u *UserAccount) setUserID() error {
	id, err := DBI.GetIntValue(fmt.Sprintf(SELECT_USERID_QUERY, u.Token.UserName))
	if err != nil {
		return err
	}
	u.Token.UserID = id
	return nil
}

// MatchPassword match password with what is in database.  Return error if passwords do not match
func (u *UserAccount) MatchPassword(p string) error {

	t, err := GetToken(u.Token.UserName)
	if err != nil {
		return err
	}

	if t.RefreshToken.String != p {
		return fmt.Errorf("Username or Password does not match")
	}
	return nil
}
