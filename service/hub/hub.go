package hub

import (
	"sync"

	"github.com/apex/log"
	"github.com/gofrs/uuid"
)

type Hub interface {
	SubscribeForOrganisation(organisationID uuid.UUID) Unsubscriber
	Publish(msg Message)
}

type hub struct {
	mu               sync.RWMutex
	organisationSubs map[uuid.UUID]map[*organisationSubscription]*organisationSubscription
}

func NewHub() Hub {
	return &hub{
		organisationSubs: make(map[uuid.UUID]map[*organisationSubscription]*organisationSubscription),
	}
}

type Message interface {
	IsMessage()
}

type OrganisationScoped interface {
	OrganisationID() uuid.UUID
}

type Unsubscriber interface {
	Unsubscribe()
}

type organisationSubscription struct {
	C              chan Message
	h              *hub
	organisationID uuid.UUID
}


func (sub *organisationSubscription) Unsubscribe() {
	sub.h.mu.Lock()
	defer sub.h.mu.Unlock()

	delete(sub.h.organisationSubs[sub.organisationID], sub)
	if len(sub.h.organisationSubs[sub.organisationID]) == 0 {
		delete(sub.h.organisationSubs, sub.organisationID)
	}

	close(sub.C)
}


func (h *hub) SubscribeForOrganisation(organisationID uuid.UUID) Unsubscriber {
	sub := &organisationSubscription{
		C:              make(chan Message),
		h:              h,
		organisationID: organisationID,
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.organisationSubs[organisationID]; !ok {
		h.organisationSubs[organisationID] = make(map[*organisationSubscription]*organisationSubscription)
	}
	h.organisationSubs[organisationID][sub] = sub

	return sub
}

func (h *hub) Publish(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	log.
		WithField("msg", msg).
		Debugf("Hub: Publishing %T message", msg)

	if orgMsg, ok := msg.(OrganisationScoped); ok {
		organisationID := orgMsg.OrganisationID()
		for sub := range h.organisationSubs[organisationID] {
			sub.C <- msg
			log.
				WithField("msg", msg).
				WithField("organisationID", sub.organisationID).
				Debugf("Hub: Published %T message to organisation subscriber", msg)
		}
	}
}
