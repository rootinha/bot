package main

type Plugin interface {
	ListActions() map[string]ConversationFn
}
