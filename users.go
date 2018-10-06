package main

import (
	"crypto/sha512"
	"database/sql"
	"log"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/unicode/norm"
)

const (
	USERNAME_MIN_LENGTH = 1
	USERNAME_MAX_LENGTH = 32
	PASSWORD_MIN_LENGTH = 8
	PASSWORD_MAX_LENGTH = 256
)

type User struct {
	Name       string
	NameNormal string
	Email      string
	Data       UserData
}

type UserData struct {
}

func preparePassword(password string) []byte {
	hashed := sha512.Sum512([]byte(password))
	return hashed[:]
}

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword(preparePassword(password), 10)
}

func comparePassword(password string, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, preparePassword(password))
}

func normalizeUsername(username string) string {
	return norm.NFKD.String(strings.ToLower(norm.NFKD.String(username)))
}

func ValidatePassword(password string) error {
	length := len([]rune(password))
	if length < PASSWORD_MIN_LENGTH {
		return ErrorPasswordTooShort
	}
	if length > PASSWORD_MAX_LENGTH {
		return ErrorPasswordTooLong
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

func CreateUser(username, password string) (User, error) {
	user := User{NameNormal: normalizeUsername(username)}
	var id int

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return user, err
	}

	err = DB.QueryRow(`INSERT INTO players(name, name_normal, password) VALUES ($1, $2, $3)
        RETURNING id`, username, user.NameNormal, hashedPassword).Scan(&id)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		return user, err
	}

	log.Printf("User created, ID: %d", id)
	user.Name = username
	return user, nil
}

func AuthenticateUser(normalizedUsername, password string) (User, error) {
	user := User{NameNormal: normalizedUsername}
	var passwordHash []byte
	var email sql.NullString

	err := DB.QueryRow(`SELECT name, password, email FROM players
        WHERE name_normal = $1`, normalizedUsername).Scan(&user.Name, &passwordHash, &email)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		if err == sql.ErrNoRows {
			return user, ErrorUserMissing
		}
		return user, err
	}
	if email.Valid {
		user.Email = email.String
	} else {
		user.Email = ""
	}

	err = comparePassword(password, passwordHash)
	if err != nil {
		log.Printf("Password Error: %v", err)
		return user, ErrorInvalidLogin
	}

	return user, nil
}

func LoadUser(username string) (User, error) {
	user := User{NameNormal: username}
	var email sql.NullString

	err := DB.QueryRow(`SELECT name, email FROM players
        WHERE name_normal = $1`, username).Scan(&user.Name, &email)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		if err == sql.ErrNoRows {
			return user, ErrorUserMissing
		}
		return user, err
	}
	if email.Valid {
		user.Email = email.String
	} else {
		user.Email = ""
	}

	return user, nil
}
