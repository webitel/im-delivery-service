package event

import "context"

type metadataKey struct{}

func ContextWithMetadata(ctx context.Context, metadata map[string]string) context.Context {
	return context.WithValue(ctx, metadataKey{}, metadata)
}

func TryGetMetadataFromContext(ctx context.Context) (map[string]string, bool) {
	md, ok := ctx.Value(metadataKey{}).(map[string]string)
	return md, ok
}
