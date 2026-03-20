package redis

import (
	goredis "github.com/redis/go-redis/v9"
	"github.com/webitel/im-delivery-service/internal/store"
	"go.uber.org/fx"
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
