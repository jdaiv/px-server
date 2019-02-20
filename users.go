package main

import (
	"database/sql"
	"log"
	"strings"
	"unicode"

	"github.com/sethvargo/go-password/password"
	"golang.org/x/text/unicode/norm"
)

const (
	USERNAME_MIN_LENGTH = 1
	USERNAME_MAX_LENGTH = 32
	PASSWORD_LENGTH     = 60
)

type User struct {
	Id         int      `json:"-"`
	Name       string   `json:"name"`
	NameNormal string   `json:"nameNormal"`
	Data       UserData `json:"userData"`
}

type UserData struct {
}

func createPassword() []byte {
	// todo: handle error
	p, _ := password.Generate(PASSWORD_LENGTH, 12, 12, true, true)
	return []byte(p)
}

func normalizeUsername(username string) string {
	return norm.NFKD.String(strings.ToLower(norm.NFKD.String(username)))
}

func ValidatePassword(password string) error {
	length := len([]rune(password))
	if length != PASSWORD_LENGTH {
		return ErrorPasswordTooShort
	}
	return nil
}

func ValidateUsername(username string) error {
	length := len([]rune(username))
	if length <= 0 {
		return ErrorUsernameTooShort
	}
	if length > USERNAME_MAX_LENGTH {
		return ErrorUsernameTooLong
	}
	for _, ch := range username {
		if unicode.IsSpace(ch) {
			return ErrorUsernameInvalidChars
		}
	}
	return nil
}

func CreateUser(username string) (User, string, error) {
	user := User{NameNormal: normalizeUsername(username)}
	password := createPassword()
	var id int

	err := DB.QueryRow(`INSERT INTO players(name, name_normal, password) VALUES ($1, $2, $3)
        RETURNING id`, username, user.NameNormal, password).Scan(&id)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		return user, "", err
	}

	log.Printf("User created, ID: %d", id)
	user.Name = username
	return user, string(password), nil
}

func AuthenticateUser(password string) (User, error) {
	user := User{}

	err := DB.QueryRow(`SELECT id, name, name_normal FROM players
        WHERE password = $1`, password).Scan(&user.Id, &user.Name, &user.NameNormal)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		if err == sql.ErrNoRows {
			return user, ErrorInvalidLogin
		}
		return user, err
	}

	return user, nil
}

func LoadUser(username string) (User, error) {
	user := User{NameNormal: username}

	err := DB.QueryRow(`SELECT id, name FROM players
        WHERE name_normal = $1`, username).Scan(&user.Id, &user.Name)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		if err == sql.ErrNoRows {
			return user, ErrorUserMissing
		}
		return user, err
	}

	return user, nil
}
