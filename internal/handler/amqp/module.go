package amqp

import (
	pubsubadapter "github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"go.uber.org/fx"
)

var Module = fx.Module("amqp-handler",
	fx.Provide(

		pubsubadapter.NewPublisherProvider,
		pubsubadapter.NewSubscriberProvider,

		NewMessageHandler,
		NewWatermillRouter,
	),

	fx.Invoke(RegisterHandlers),
)
