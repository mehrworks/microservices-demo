package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseFrontendMode(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want frontendMode
	}{
		{name: "default empty", raw: "", want: frontendModeFull},
		{name: "full", raw: "full", want: frontendModeFull},
		{name: "catalog only", raw: "catalog-only", want: frontendModeCatalogOnly},
		{name: "catalog only normalized", raw: "  CATALOG-ONLY  ", want: frontendModeCatalogOnly},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseFrontendMode(tc.raw); got != tc.want {
				t.Fatalf("parseFrontendMode(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseFrontendModePanicsOnInvalidValue(t *testing.T) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatal("parseFrontendMode did not panic for invalid mode")
		}
	}()

	_ = parseFrontendMode("unsupported")
}

func TestCurrentCurrencyIgnoresCookieInCatalogOnlyMode(t *testing.T) {
	previousMode := activeFrontendMode
	activeFrontendMode = frontendModeCatalogOnly
	t.Cleanup(func() {
		activeFrontendMode = previousMode
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieCurrency, Value: "EUR"})

	if got := currentCurrency(req); got != defaultCurrency {
		t.Fatalf("currentCurrency() = %q, want %q", got, defaultCurrency)
	}
}

func TestCurrentCurrencyRespectsCookieInFullMode(t *testing.T) {
	previousMode := activeFrontendMode
	activeFrontendMode = frontendModeFull
	t.Cleanup(func() {
		activeFrontendMode = previousMode
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieCurrency, Value: "EUR"})

	if got := currentCurrency(req); got != "EUR" {
		t.Fatalf("currentCurrency() = %q, want %q", got, "EUR")
	}
}
