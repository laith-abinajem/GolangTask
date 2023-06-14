package model

import "sync"

type Account struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Balance string `json:"balance"`
}

type Accounts struct {
	sync.RWMutex
	Data []Account
}
