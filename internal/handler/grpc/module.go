// internal/handler/grpc/module.go
package grpc

import (
	"go.uber.org/fx"

	impb "github.com/webitel/im-delivery-service/gen/go/api/v1"
	grpcsrv "github.com/webitel/im-delivery-service/infra/server/grpc"
)

var Module = fx.Module("delivery-grpc",
	fx.Provide(
		NewDeliveryService,
	),
	fx.Invoke(RegisterDeliveryServices),
)

func RegisterDeliveryServices(
	server *grpcsrv.Server,
	service *DeliveryService,
) {
	impb.RegisterDeliveryServer(server.Server, service)
}
