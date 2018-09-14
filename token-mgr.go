package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	tokenTable = "users"
)

// EncryptedText convert string into encrypted form
type EncryptedText struct {
	String string
}

// TokenTuple represents database row in the users table
type TokenTuple struct {
	UserID       int           `json:"userid"`
	UserName     string        `json:"username"`
	RefreshToken EncryptedText `json:"refreshtoken"`
}

// NoTokenFound returned when unable to retrieve token from database
var NoTokenFound = errors.New("Token Manager: No Refresh Token Found")

// Value encrypt data going into database
func (et EncryptedText) Value() (driver.Value, error) {
	gcm, err := newGCM(EncryptionKey)
	if err != nil {
		return driver.Value(""), fmt.Errorf("CIPHER ERROR: %s", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return driver.Value(""), fmt.Errorf("Could not generated new gcm random nonce: %s", err)
	}
	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(et.String), nil)), nil
}

// Scan decrypt data from database
func (et *EncryptedText) Scan(value interface{}) error {
	if value == nil {
		et.String = ""
		return nil
	}

	ciphertext, err := hex.DecodeString(fmt.Sprintf("%s", value))
	if err != nil {
		return fmt.Errorf("Could not decode string: %s", err)
	}
	gcm, err := newGCM(EncryptionKey)
	if err != nil {
		return fmt.Errorf("CIPHER ERROR: %s", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		fmt.Printf("et=%s\nciphertext=%d\nnonceSize=%d\n", et.String, len(ciphertext), nonceSize)
		return fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}
	et.String = fmt.Sprintf("%s", decrypted)
	return nil
}

// create a new gcm and return
func newGCM(key string) (cipher.AEAD, error) {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("Could not generated new cipher: %s", err)
	}
	return cipher.NewGCM(c)
}

/*
	Fetch the user refresh token from database
	if no token found then return NoTokenFound
*/
func GetToken(u string) (*TokenTuple, error) {

	t := new(TokenTuple)
	q := fmt.Sprintf("SELECT * FROM %s WHERE username='%s' limit 1", tokenTable, u)
	err := DBI.SQLSession.QueryRow(q).Scan(&t.UserID, &t.UserName, &t.RefreshToken)
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

	tNew, err := GetToken(t.UserName)
	if err == NoTokenFound {
		dberr = DBI.ExecTXQuery("INSERT INTO users values(DEFAULT, ?,?)", t.UserName, t.RefreshToken)
	} else {
		dberr = DBI.ExecTXQuery("UPDATE user SET username = ?, refreshtoken = ? WHERE id = ?", t.UserName, t.RefreshToken, tNew.UserID)
	}

	if dberr != nil {
		fmt.Printf("UpdateToken() Error when running query %s: %s\n", q, dberr)
		return dberr
	}
	return nil
}
