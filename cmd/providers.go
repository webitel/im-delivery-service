package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/resource"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.38.0"
	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/infra/discovery"
	otelsdk "github.com/webitel/webitel-go-kit/infra/otel/sdk"
	"github.com/webitel/webitel-go-kit/infra/profiler"
	"github.com/webitel/webitel-go-kit/pkg/depenlog"
	"github.com/webitel/webitel-go-kit/pkg/logger"
	"github.com/webitel/webitel-go-kit/pkg/semconv"

	"github.com/webitel/im-delivery-service/config"
	"github.com/webitel/im-delivery-service/infra/pubsub"
	"github.com/webitel/im-delivery-service/infra/pubsub/factory"
	"github.com/webitel/im-delivery-service/infra/pubsub/factory/amqp"
	"github.com/webitel/im-delivery-service/internal/domain/model"

	_ "github.com/webitel/webitel-go-kit/infra/discovery/consul"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/log/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/log/stdout"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/metric/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/metric/stdout"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/trace/otlp"
	_ "github.com/webitel/webitel-go-kit/infra/otel/sdk/trace/stdout"
)

func ProvideWatermillLogger(l *slog.Logger) watermill.LoggerAdapter {
	return watermill.NewSlogLogger(l)
}

func ProvideLogger(cfg *config.Config, lc fx.Lifecycle) (*slog.Logger, logger.Logger, error) {
	logSettings := cfg.Log

	if !logSettings.Console && !logSettings.Otel && logSettings.File == "" {
		logSettings.Console = true
	}

	dcfg := depenlog.Config{
		Level:   logSettings.Level,
		JSON:    logSettings.JSON,
		File:    logSettings.File,
		Console: logSettings.Console,
	}

	var opts []depenlog.Option

	if logSettings.Otel {
		service := resource.NewSchemaless(
			otelsemconv.ServiceName(model.ServiceName),
			otelsemconv.ServiceVersion(model.Version),
			otelsemconv.ServiceInstanceID(discovery.GenerateInstanceID(model.ServiceName)),
			otelsemconv.ServiceNamespace(model.ServiceNamespace),
		)
		otelHandler := otelslog.NewHandler("slog")

		shutdown, err := otelsdk.Configure(context.Background(), otelsdk.WithResource(service),
			otelsdk.WithLogBridge(
				func() {
					opts = append(opts, depenlog.WithHandler(otelHandler))
				},
			),
		)
		if err != nil {
			return nil, nil, err
		}

		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return shutdown(ctx)
			},
		})
	}

	kit := depenlog.New(dcfg, opts...)

	return slog.Default(), kit, nil
}

func ProvideSD(cfg *config.Config, log *slog.Logger, lc fx.Lifecycle) (discovery.DiscoveryProvider, error) {
	provider, err := discovery.DefaultFactory.CreateProvider(
		discovery.ProviderConsul,
		log,
		cfg.Consul.Addr,
		discovery.WithHeartbeat[discovery.DiscoveryProvider](true),
		discovery.WithTimeout[discovery.DiscoveryProvider](time.Second*30),
	)
	if err != nil {
		return nil, err
	}

	si := new(discovery.ServiceInstance)
	{
		si.Id = discovery.GenerateInstanceID(model.ServiceName)
		si.Name = model.ServiceName
		si.Version = model.Version
		si.Metadata = map[string]string{
			"commit":         model.Commit,
			"commitDate":     model.CommitDate,
			"branch":         model.Branch,
			"buildTimestamp": model.BuildTimestamp,
		}
		si.Endpoints = []string{(&url.URL{Scheme: "grpc", Host: cfg.Service.Addr}).String()}
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := provider.Register(ctx, si); err != nil {
				return err
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := provider.Deregister(ctx, si); err != nil {
				return err
			}

			return nil
		},
	})

	return provider, nil
}

func ProvidePubSub(cfg *config.Config, l *slog.Logger, lc fx.Lifecycle) (pubsub.Provider, error) {
	var (
		pubsubConfig  = cfg.Pubsub
		loggerAdapter = watermill.NewSlogLogger(l)
		pubsubFactory factory.Factory
		err           error
	)

	switch pubsubConfig.Driver {
	case "amqp":
		pubsubFactory, err = amqp.NewFactory(pubsubConfig.URL, loggerAdapter)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("pubsub driver not supported")
	}

	router, err := message.NewRouter(message.RouterConfig{}, loggerAdapter)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := router.Run(context.Background()); err != nil {
					l.Error("watermill router failed", slog.Any(semconv.ErrorKey, err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return router.Close()
		},
	})

	return pubsub.NewDefaultProvider(router, pubsubFactory)
}

func ProvideRedis(cfg *config.Config, lc fx.Lifecycle, l *slog.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := rdb.Ping(ctx).Err()
			if err != nil {
				l.Error("redis connection failed", slog.Any(semconv.ErrorKey, err))

				return err
			}

			l.Info("redis connected", slog.String("addr", cfg.Redis.Addr))

			return nil
		},
		OnStop: func(ctx context.Context) error {
			l.Info("closing redis connection")

			return rdb.Close()
		},
	})

	return rdb, nil
}

func ProvideProfiler(cfg *config.Config) profiler.Config {
	return profiler.Config{
		Addr:                 cfg.Profiler.Addr,
		MutexProfileFraction: cfg.Profiler.MutexFraction,
		BlockProfileRate:     cfg.Profiler.BlockRate,
	}
}
