package payload

import (
	"strconv"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

type Peer struct {
	ID   string `json:"id"`
	Type int    `json:"type"`
}

type MessageCreatedV1 struct {
	MessageID  string     `json:"message_id"`
	ThreadID   string     `json:"thread_id"`
	DomainID   int32      `json:"domain_id"`
	From       Peer       `json:"from"`
	To         []string   `json:"to"`
	Body       string     `json:"body"`
	OccurredAt string     `json:"occurred_at"`
	SendID     string     `json:"send_id"`
	Images     []Image    `json:"images"`
	Documents  []Document `json:"documents"`
}

func (d *MessageCreatedV1) ToDomain() *model.Message {
	return &model.Message{
		ID:        util.SafeParseUUID(d.MessageID),
		SendID:    d.SendID,
		ThreadID:  util.SafeParseUUID(d.ThreadID),
		DomainID:  int64(d.DomainID),
		Text:      d.Body,
		CreatedAt: util.SafeParseRFC3339(d.OccurredAt),
		Images:    d.mapImages(),
		Documents: d.mapDocs(),
		Metadata:  make(map[string]any),
	}
}

func (d Peer) ToDomain() model.Peer {
	return model.NewPeer(
		util.SafeParseUUID(d.ID),
		model.PeerType(d.Type),
	)
}

func (d *MessageCreatedV1) mapImages() []*model.Image {
	res := make([]*model.Image, 0, len(d.Images))
	for _, img := range d.Images {
		res = append(res, &model.Image{
			ID:       strconv.FormatInt(img.FileID, 10),
			FileName: img.Name,
			MimeType: img.Mime,
			URL:      img.URL,
		})
	}
	return res
}

func (d *MessageCreatedV1) mapDocs() []*model.Document {
	res := make([]*model.Document, 0, len(d.Documents))
	for _, doc := range d.Documents {
		res = append(res, &model.Document{
			ID:       strconv.FormatInt(doc.FileID, 10),
			FileName: doc.Name,
			MimeType: doc.Mime,
			Size:     doc.Size,
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
}
