package constants

type MessageType int8

const (
	MessageTypeRealExchange MessageType = 2
	MessageTypeRequester    MessageType = 1
	MessageTypeListener     MessageType = 0
	MessageTypeFailure      MessageType = -1
	MessageTypeFakeChatter  MessageType = -2
)
