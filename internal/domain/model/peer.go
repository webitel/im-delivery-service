package model

import (
	"strings"

	"github.com/google/uuid"
)

//go:generate stringer -type=PeerType
type PeerType int16

const (
	PeerUser PeerType = iota + 1
	PeerGroup
	PeerChannel
)

// ParsePeerType maps a contact's channel type onto the addressing kind. Every
// unrecognised type addresses a user: PeerType has no zero member, and an unset
// one leaves a gRPC Peer without any Kind.
func ParsePeerType(t string) PeerType {
	switch t {
	case "group":
		return PeerGroup
	case "channel":
		return PeerChannel
	default:
		return PeerUser
	}
}

type Peer struct {
	ID          uuid.UUID `json:"id"`
	Type        PeerType  `json:"type"`
	ContactType string    `json:"-"`
	Sub         string    `json:"sub,omitempty"`
	Issuer      string    `json:"issuer,omitempty"`
	Name        string    `json:"name,omitempty"`
	Username    string    `json:"username,omitempty"`
	IsBot       bool      `json:"is_bot"`
	MemberID    string    `json:"member_id,omitempty"`
	Role        int32     `json:"role,omitempty"`
	DomainID    int       `json:"domain_id,omitempty"`
}

type PeerOption func(*Peer)

// WithIdentity applies enrichment data from external services.
func WithIdentity(sub, issuer, name string) PeerOption {
	return func(p *Peer) {
		p.Sub = sub
		p.Issuer = issuer
		p.Name = name
	}
}

// [OPTION] WithBot sets the bot status for the peer.
func WithBot(isBot bool) PeerOption {
	return func(p *Peer) {
		p.IsBot = isBot
	}
}

func WithDomainID(domainID int) PeerOption { return func(p *Peer) { p.DomainID = domainID } }

func NewPeer(id uuid.UUID, pType PeerType, opts ...PeerOption) Peer {
	p := Peer{ID: id, Type: pType}
	for _, opt := range opts {
		opt(&p)
	}

	return p
}

// IsEnriched determines if the peer has verified identity metadata.
func (p Peer) IsEnriched() bool {
	return p.Sub != ""
}

// GetRoutingParts returns normalized segments for RabbitMQ routing keys.
func (p Peer) GetRoutingParts() (sub, issuer string) {
	sub, issuer = "any", "any"
	// [NORMALIZATION] Ensure consistent casing for routing logic.
	if p.Sub != "" {
		sub = strings.ToLower(p.Sub)
	}

	if p.Issuer != "" {
		issuer = strings.ToLower(p.Issuer)
	}

	return sub, issuer
}
