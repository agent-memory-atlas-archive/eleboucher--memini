package embed_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/eleboucher/memini/internal/embed"
)

func TestDocumentPrefixSeparatesCachedQueries(t *testing.T) {
	inner := &countingEmbedder{dims: 2}
	cached, err := embed.NewCached(inner, 8)
	if err != nil {
		t.Fatal(err)
	}
	e := embed.WithDocumentPrefix(cached, "document: ")
	ctx := context.Background()
	texts := []string{"same", "other"}
	for range 2 {
		if _, err := e.Embed(ctx, texts); err != nil {
			t.Fatal(err)
		}
		if _, err := embed.EmbedQueryOne(ctx, e, "same"); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(texts, []string{"same", "other"}) {
		t.Fatalf("caller inputs mutated: %q", texts)
	}
	if want := []string{"document: same", "document: other", "same"}; !reflect.DeepEqual(inner.seen, want) {
		t.Fatalf("endpoint inputs = %q, want %q", inner.seen, want)
	}
}
