package main

import (
	"crypto/sha512"
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name  string
	Email string
	Data  UserData
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

func CreateUser(username, password string) (User, error) {
	user := User{}
	var id int

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return user, err
	}

	err = DB.QueryRow(`INSERT INTO players(name, password) VALUES ($1, $2)
        RETURNING id`, username, hashedPassword).Scan(&id)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		return user, err
	}

	log.Printf("User created, ID: %d", id)
	user.Name = username
	return user, nil
}

func AuthenticateUser(username, password string) (User, error) {
	user := User{}
	var passwordHash []byte
	var email sql.NullString

	err := DB.QueryRow(`SELECT name, password, email FROM players
        WHERE name = $1`, username).Scan(&user.Name, &passwordHash, &email)
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
		return user, ErrorInvalidPassword
	}

	return user, nil
}

func LoadUser(username string) (User, error) {
	user := User{}
	var email sql.NullString

	err := DB.QueryRow(`SELECT name, email FROM players
        WHERE name = $1`, username).Scan(&user.Name, &email)
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
