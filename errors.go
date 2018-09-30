package main

import (
	"fmt"
)

var (
	ErrorInternal    = NewClientError(1, "internal error", "internal error")
	ErrorInvalidData = NewClientError(2, "invalid data", "invalid data")

	ErrorMissingAction = NewClientError(1001,
		"ws: missing action", "missing action")
	ErrorMissingScope = NewClientError(1002,
		"ws: missing scope", "missing scope")

	ErrorInvalidPassword = NewClientError(3001,
		"auth: invalid password", "invalid password")
	ErrorUserMissing = NewClientError(3002,
		"auth: user does not exist", "user not found")

	ErrorInvalidToken = NewClientError(3101,
		"ws_auth: invalid token", "invalid token")
	ErrorUnauthenticated = NewClientError(3102,
		"ws_auth: unauthenticated", "unauthenticated")

	ErrorRoomExists = NewClientError(2001,
		"rooms: room already exists", "room already exists")
	ErrorRoomMissing = NewClientError(2002,
		"rooms: room does not exist", "room does not exist")
	ErrorAlreadyInRoom = NewClientError(2003,
		"rooms: client already in room", "client already in room")
	ErrorClientHasRoom = NewClientError(2004,
		"rooms: client already has a room", "client already has a room")
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
