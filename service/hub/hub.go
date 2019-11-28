package hub

import (
	"sync"

	"github.com/apex/log"
	"github.com/gofrs/uuid"
)

type Hub struct {
	mu               sync.RWMutex
	organisationSubs map[uuid.UUID]map[*OrganisationSubscription]*OrganisationSubscription
}

func NewHub() *Hub {
	return &Hub{
		organisationSubs: make(map[uuid.UUID]map[*OrganisationSubscription]*OrganisationSubscription),
	}
}

type Message interface {
	IsMessage()
}

type OrganisationScoped interface {
	OrganisationID() uuid.UUID
}

type OrganisationSubscription struct {
	C              chan Message
	h              *Hub
	organisationID uuid.UUID
}


func (sub *OrganisationSubscription) Unsubscribe() {
	sub.h.mu.Lock()
	defer sub.h.mu.Unlock()

	delete(sub.h.organisationSubs[sub.organisationID], sub)
	if len(sub.h.organisationSubs[sub.organisationID]) == 0 {
		delete(sub.h.organisationSubs, sub.organisationID)
	}

	close(sub.C)
}


func (h *Hub) SubscribeForOrganisation(organisationID uuid.UUID) *OrganisationSubscription {
	sub := &OrganisationSubscription{
		C:              make(chan Message),
		h:              h,
		organisationID: organisationID,
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.organisationSubs[organisationID]; !ok {
		h.organisationSubs[organisationID] = make(map[*OrganisationSubscription]*OrganisationSubscription)
	}
	h.organisationSubs[organisationID][sub] = sub

	return sub
}

func (h *Hub) Publish(msg Message) {
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
