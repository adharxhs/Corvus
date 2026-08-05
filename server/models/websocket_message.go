package models

type ConnectionState int

const (
	StateConnected ConnectionState = iota
	StateDisconnected
)
