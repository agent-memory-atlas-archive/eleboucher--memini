package embed

import "context"

type queryKey struct{}

// EmbedQueryOne embeds a recall query without the document prefix.
func EmbedQueryOne(ctx context.Context, e Embedder, text string) ([]float32, error) {
	return EmbedOne(context.WithValue(ctx, queryKey{}, true), e, text)
}

type prefixed struct {
	Embedder
	prefix string
}

// WithDocumentPrefix prepends prefix to documents, leaving recall queries alone.
// Wrap outside caches so cache keys include the prefix and distinguish queries.
func WithDocumentPrefix(inner Embedder, prefix string) Embedder {
	if prefix == "" {
		return inner
	}
	return &prefixed{Embedder: inner, prefix: prefix}
}

func (p *prefixed) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if query, _ := ctx.Value(queryKey{}).(bool); query {
		return p.Embedder.Embed(ctx, texts)
	}
	inputs := make([]string, len(texts))
	for i, text := range texts {
		inputs[i] = p.prefix + text
	}
	return p.Embedder.Embed(ctx, inputs)
}
