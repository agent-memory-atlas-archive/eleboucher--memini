package llm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eleboucher/memini/internal/llm"
)

func TestDistillModelRouting(t *testing.T) {
	for _, api := range []llm.API{llm.APIOpenAI, llm.APIAnthropic} {
		for _, override := range []string{"", "small-model"} {
			t.Run(fmt.Sprintf("%s/override=%s", api, override), func(t *testing.T) {
				wantModel := "large-model"
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body["model"] != wantModel {
						t.Errorf("model = %v, want %s", body["model"], wantModel)
					}
					if body["max_tokens"] != float64(1234) || body["custom"] != "shared" {
						t.Errorf("shared options missing: %v", body)
					}
					if r.Header.Get("X-Routing") != "shared" {
						t.Error("missing shared header")
					}
					if api == llm.APIOpenAI && r.Header.Get("Authorization") != "Bearer secret" {
						t.Error("missing API key")
					}
					if api == llm.APIAnthropic && r.Header.Get("X-Api-Key") != "secret" {
						t.Error("missing API key")
					}
					w.Header().Set("Content-Type", "application/json")
					content, _ := json.Marshal(`{"facts":[],"action":"new"}`)
					if api == llm.APIAnthropic {
						_, _ = fmt.Fprintf(w, `{"id":"msg_1","type":"message","role":"assistant","model":"m","content":[{"type":"text","text":%s}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`, content)
					} else {
						_, _ = fmt.Fprintf(w, `{"choices":[{"message":{"role":"assistant","content":%s}}]}`, content)
					}
				}))
				defer srv.Close()
				c, err := llm.New(api, llm.Config{BaseURL: srv.URL, APIKey: "secret", Model: "large-model", DistillModel: override, MaxTokens: 1234,
					ExtraHeaders: map[string]string{"X-Routing": "shared"}, ExtraBody: map[string]json.RawMessage{"custom": json.RawMessage(`"shared"`)}})
				if err != nil {
					t.Fatal(err)
				}
				ctx := context.Background()
				if _, err := c.Complete(ctx, "sys", "user"); err != nil {
					t.Fatal(err)
				}
				if override != "" {
					wantModel = override
				}
				if _, err := c.Distill(ctx, llm.DistillInput{}); err != nil {
					t.Fatal(err)
				}
				wantModel = "large-model"
				if _, err := c.Complete(ctx, "sys", "user"); err != nil {
					t.Fatal(err)
				}
				if _, err := c.Consolidate(ctx, llm.Input{New: "fact"}); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
