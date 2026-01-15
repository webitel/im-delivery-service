package pubsub

import (
	"github.com/ThreeDotsLabs/watermill/message"
	infrapubsub "github.com/webitel/im-delivery-service/infra/pubsub"
	"github.com/webitel/im-delivery-service/infra/pubsub/factory"
)

type PublisherProvider struct {
	factory factory.Factory
}

func NewPublisherProvider(p infrapubsub.Provider) *PublisherProvider {
	return &PublisherProvider{factory: p.GetFactory()}
}

func (pp *PublisherProvider) Build(exchange string) (message.Publisher, error) {
	return pp.factory.BuildPublisher(&factory.PublisherConfig{
		Exchange: factory.ExchangeConfig{
			Name:    exchange,
			Type:    "topic",
			Durable: true,
		},
	})
}
