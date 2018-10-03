package main

import (
	"fmt"
)

type WSHandler func(source *Client, target string, data []byte) (interface{}, error)

// action - scope/type/target

type WSRouter struct {
	Handlers map[string]map[string]WSHandler
}

func NewWSRouter() *WSRouter {
	return &WSRouter{
		Handlers: make(map[string]map[string]WSHandler),
	}
}

func (r *WSRouter) AddHandler(scope, actionType string, h WSHandler) error {
	scopeMap, hasScope := r.Handlers[scope]
	if !hasScope {
		scopeMap = make(map[string]WSHandler)
		r.Handlers[scope] = scopeMap
	}
	if _, hasAction := scopeMap[actionType]; hasAction {
		return fmt.Errorf("wsrouters: route %s/%s already exists", scope, actionType)
	}
	scopeMap[actionType] = h
	return nil
}

func (r *WSRouter) GetHandler(scope, actionType string) (WSHandler, error) {
	scopeMap, hasScope := r.Handlers[scope]
	if !hasScope {
		return nil, fmt.Errorf("wsrouters: scope %s doesn't exist", scope)
	}
	if _, hasAction := scopeMap[actionType]; !hasAction {
		return nil, fmt.Errorf("wsrouters: route %s/%s doesn't exist", scope, actionType)
	}
	return scopeMap[actionType], nil
}
