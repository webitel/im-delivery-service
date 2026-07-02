package redis

import (
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/webitel/im-delivery-service/internal/store"
)

var Module = fx.Options(
	fx.Provide(
		func(rdb *goredis.Client) store.PresenceStore {
			return NewRedisPresence(rdb)
		},
		func(rdb *goredis.Client) store.DeliveryTracker {
			return NewRedisTracker(rdb)
		},
		func(rdb *goredis.Client) store.DeliveryScheduler {
			return NewRedisScheduler(rdb)
		},
	),
)
