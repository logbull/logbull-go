package core

import (
	"sync"
)

var (
	senderRegistry = &registry{
		senders: make([]*Sender, 0),
	}
)

type registry struct {
	mu      sync.Mutex
	senders []*Sender
}

func (r *registry) register(sender *Sender) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.senders = append(r.senders, sender)
}

func registerSender(sender *Sender) {
	senderRegistry.register(sender)
}
