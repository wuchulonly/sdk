package neutron

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainreactors/sdk/pkg/types"
)

func TestExecuteWithTransportPreservesRedirectPolicy(t *testing.T) {
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/redirect") {
			redirectCount++
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("final"))
	}))
	defer server.Close()

	tpl := parseTemplateForTest(t, `id: transport-redirect
info:
  name: Transport Redirect Test
  severity: info
http:
  - method: GET
    path:
      - "{{BaseURL}}/redirect"
    redirects: true
    max-redirects: 3
    matchers:
      - type: word
        words:
          - "final"
`)

	engine := &Engine{config: NewConfig()}
	compiled := engine.compileTemplates([]*types.Template{tpl})
	if len(compiled) != 1 {
		t.Fatalf("expected 1 compiled template, got %d", len(compiled))
	}

	result, err := compiled[0].ExecuteWithTransport(server.URL, nil, http.DefaultTransport)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result == nil || !result.Matched {
		t.Fatal("expected template to follow redirect and match 'final'")
	}
}
