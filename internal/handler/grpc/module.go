// internal/handler/grpc/module.go
package grpc

import (
	"go.uber.org/fx"

	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	grpcsrv "github.com/webitel/im-delivery-service/infra/server/grpc"
	grpcmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/gprc"
)

var Module = fx.Module("delivery-grpc",
	fx.Provide(
		grpcmarshaller.New,
		NewDeliveryHandler,
	),
	fx.Invoke(RegisterDeliveryServices),
)

func RegisterDeliveryServices(
	server *grpcsrv.Server,
	service *DeliveryHandler,
) {
	impb.RegisterDeliveryServer(server.Server, service)
}
