package main

import (
	"fmt"
)

var (
	ErrorInternal = NewClientError(1,
		"internal error", "internal error")
	ErrorInvalidData = NewClientError(2,
		"invalid data", "invalid data")

	ErrorMissingAction = NewClientError(1001,
		"ws: missing action", "missing action")
	ErrorMissingScope = NewClientError(1002,
		"ws: missing scope", "missing scope")

	ErrorInvalidLogin = NewClientError(3001,
		"auth: invalid login", "invalid username and password combination")
	ErrorUserMissing = NewClientError(3002,
		"auth: user does not exist", "user not found")

	ErrorUsernameTooShort = NewClientError(3101,
		"auth: username too short", "username too short")
	ErrorUsernameTooLong = NewClientError(3102,
		"auth: username too long", "username too long")
	ErrorUsernameInvalidChars = NewClientError(3103,
		"auth: username contains invalid characters", "username contains invalid characters")

	ErrorPasswordTooShort = NewClientError(3201,
		"auth: password too short", "password too short")
	ErrorPasswordTooLong = NewClientError(3202,
		"auth: password too long", "password too long")

	ErrorInvalidToken = NewClientError(3101,
		"ws_auth: invalid token", "invalid token")
	ErrorUnauthenticated = NewClientError(3102,
		"ws_auth: unauthenticated", "unauthenticated")

	ErrorRoomExists = NewClientError(2001,
		"rooms: room already exists", "room already exists")
	ErrorRoomMissing = NewClientError(2002,
		"rooms: room does not exist", "room does not exist")
	ErrorAlreadyInRoom = NewClientError(2003,
		"rooms: client already in room", "you're already in room")
	ErrorClientHasRoom = NewClientError(2004,
		"rooms: client already has a room", "you already have a room")
	ErrorWrongRoom = NewClientError(2005,
		"rooms: client doesn't belong to room", "you're not in this room")
	ErrorNotOwner = NewClientError(2006,
		"rooms: client isn't owner", "you're not the owner")

	ErrorActMissing = NewClientError(4001,
		"activities: doesn't exist", "activity doesn't exist")
	ErrorActInvalidAction = NewClientError(4002,
		"activities: invalid action", "invalid action")
	ErrorActError = NewClientError(4003,
		"activities: encountered an error", "activity encountered an error")
)

type ClientError interface {
	error
	Code() int
	InternalMessage() string
	ExternalMessage() string
}

type clientErrorData struct {
	code            int
	internalMessage string
	externalMessage string
	origError       error
}

func NewClientError(code int, internal, external string) ClientError {
	return &clientErrorData{
		code:            code,
		internalMessage: internal,
		externalMessage: external,
	}
}

func (c clientErrorData) Error() string {
	return fmt.Sprintf("(%d) %s", c.code, c.internalMessage)
}

func (c clientErrorData) Code() int {
	return c.code
}

func (c clientErrorData) InternalMessage() string {
	return c.internalMessage
}

func (c clientErrorData) ExternalMessage() string {
	return c.externalMessage
}
