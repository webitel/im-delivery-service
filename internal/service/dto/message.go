// internal/service/dto/message.go
package dto

import (
	"strconv"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

type PeerDTO struct {
	ID   string `json:"id"`
	Type int    `json:"type"`
}

type MessageV1 struct {
	MessageID  string        `json:"message_id"`
	ThreadID   string        `json:"thread_id"`
	DomainID   int32         `json:"domain_id"`
	From       PeerDTO       `json:"from"`
	To         PeerDTO       `json:"to"`
	Body       string        `json:"body"`
	OccurredAt string        `json:"occurred_at"`
	Images     []ImageDTO    `json:"images"`
	Documents  []DocumentDTO `json:"documents"`
}

func (d *MessageV1) ToDomain() *model.Message {
	return &model.Message{
		ID:        util.SafeParseUUID(d.MessageID),
		ThreadID:  util.SafeParseUUID(d.ThreadID),
		Text:      d.Body,
		CreatedAt: util.SafeParseRFC3339(d.OccurredAt),
		Images:    d.mapImages(),
		Documents: d.mapDocs(),
		Metadata:  make(map[string]any),
	}
}

func (d PeerDTO) ToDomain() model.Peer {
	return model.NewPeer(
		util.SafeParseUUID(d.ID),
		model.PeerType(d.Type),
	)
}

func (d *MessageV1) mapImages() []*model.Image {
	res := make([]*model.Image, 0, len(d.Images))
	for _, img := range d.Images {
		res = append(res, &model.Image{
			ID:       strconv.FormatInt(img.FileID, 10),
			FileName: img.Name,
			MimeType: img.Mime,
		})
	}
	return res
}

func (d *MessageV1) mapDocs() []*model.Document {
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

type ImageDTO struct {
	FileID int64  `json:"file_id"`
	Mime   string `json:"mime"`
	Name   string `json:"name"`
}

type DocumentDTO struct {
	FileID int64  `json:"file_id"`
	Mime   string `json:"mime"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
}
