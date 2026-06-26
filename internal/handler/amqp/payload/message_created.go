package payload

import (
	"encoding/json"
	"maps"
	"strconv"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// Peer represents the sender's data in the incoming payload.
type Peer struct {
	ContactID string `json:"contact_id"`
	MemberID  string `json:"member_id"`
	Role      int    `json:"role"`
	Type      int    `json:"type"`
}

// Recipient represents a target participant in the conversation.
type Recipient struct {
	MemberID  string `json:"member_id"`
	ContactID string `json:"contact_id"`
	Role      int    `json:"role"`
}

// MessageCreatedV1 is the top-level structure for the version 1 message event.
type MessageCreatedV1 struct {
	MessageID   string          `json:"message_id"`
	ThreadID    string          `json:"thread_id"`
	DomainID    int32           `json:"domain_id"`
	From        Peer            `json:"from"`
	To          []Recipient     `json:"to"`
	Body        string          `json:"body"`
	OccurredAt  string          `json:"occurred_at"`
	SendID      string          `json:"send_id"`
	Images      []Image         `json:"images"`
	Documents   []Document      `json:"documents"`
	Interactive json.RawMessage `json:"interactive,omitempty"`
	Location    *Location       `json:"location,omitempty"`
	Contact     *Contact        `json:"contact,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	System      *System         `json:"system,omitempty"`
}

// ToDomain converts the AMQP payload into the internal domain model.
func (d *MessageCreatedV1) ToDomain() *model.Message {
	// Map the basic message structure and sender info.
	msg := &model.Message{
		ID:        util.SafeParseUUID(d.MessageID),
		SendID:    d.SendID,
		ThreadID:  util.SafeParseUUID(d.ThreadID),
		DomainID:  int64(d.DomainID),
		Text:      d.Body,
		CreatedAt: util.SafeParseRFC3339(d.OccurredAt),
		Images:    d.mapImages(),
		Documents: d.mapDocs(),
		Metadata:  d.Metadata,
		// Populate the sender with provided contact/member details.
		From: model.Peer{
			ID:       util.SafeParseUUID(d.From.ContactID),
			MemberID: d.From.MemberID,
			Role:     int32(d.From.Role),
		},
		Interactive: d.Interactive,
		Location:    d.Location.AsModel(),
		Contact:     d.Contact.AsModel(),
		System:      d.System.AsModel(),
	}

	// Map all recipients into the domain model.
	msg.To = make([]model.Peer, 0, len(d.To))
	for _, r := range d.To {
		msg.To = append(msg.To, model.Peer{
			ID:       util.SafeParseUUID(r.ContactID),
			MemberID: r.MemberID,
			Role:     int32(r.Role),
		})
	}

	return msg
}

// mapImages converts payload image entries into domain model images.
func (d *MessageCreatedV1) mapImages() []*model.Image {
	res := make([]*model.Image, 0, len(d.Images))
	for _, img := range d.Images {
		res = append(res, &model.Image{
			ID:   strconv.FormatInt(img.FileID, 10),
			Name: img.Name,
			Mime: img.Mime,
			URL:  img.URL,
		})
	}

	return res
}

// mapDocs converts payload document entries into domain model documents.
func (d *MessageCreatedV1) mapDocs() []*model.Document {
	res := make([]*model.Document, 0, len(d.Documents))
	for _, doc := range d.Documents {
		res = append(res, &model.Document{
			ID:   strconv.FormatInt(doc.FileID, 10),
			Name: doc.Name,
			Mime: doc.Mime,
			Size: doc.Size,
			URL:  doc.URL,
		})
	}

	return res
}

type Image struct {
	FileID int64  `json:"file_id"`
	Mime   string `json:"mime"`
	Name   string `json:"name"`
	URL    string `json:"url"`
}

type Document struct {
	FileID int64  `json:"file_id"`
	Mime   string `json:"mime"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	URL    string `json:"url"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   *string `json:"address,omitempty"`
	Name      *string `json:"name,omitempty"`
}

func (location *Location) AsModel() *model.Location {
	if location == nil {
		return nil
	}

	return &model.Location{
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Address:   location.Address,
		Name:      location.Name,
	}
}

type Contact struct {
	Name  *string `json:"name,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty"`
}

func (contact *Contact) AsModel() *model.Contact {
	if contact == nil {
		return nil
	}

	return &model.Contact{
		Name:  contact.Name,
		Phone: contact.Phone,
		Email: contact.Email,
	}
}

type System struct {
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (system *System) AsModel() *model.System {
	if system == nil {
		return nil
	}

	return &model.System{
		Type:     system.Type,
		Metadata: maps.Clone(system.Metadata),
	}
}
