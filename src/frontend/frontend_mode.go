package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/genproto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type frontendMode string

const (
	frontendModeEnvVar      = "FRONTEND_MODE"
	frontendModeFull        = frontendMode("full")
	frontendModeCatalogOnly = frontendMode("catalog-only")

	catalogOnlyModeNotice = "Catalog-only mode is active. Product browsing is available here while cart, checkout, and other multi-service flows stay disabled."
)

var activeFrontendMode = frontendModeFull

func loadFrontendModeFromEnv() frontendMode {
	return parseFrontendMode(os.Getenv(frontendModeEnvVar))
}

func parseFrontendMode(raw string) frontendMode {
	switch mode := frontendMode(strings.TrimSpace(strings.ToLower(raw))); mode {
	case "", frontendModeFull:
		return frontendModeFull
	case frontendModeCatalogOnly:
		return frontendModeCatalogOnly
	default:
		panic(fmt.Sprintf("unsupported %s %q; expected %q or %q", frontendModeEnvVar, raw, frontendModeFull, frontendModeCatalogOnly))
	}
}

func catalogOnlyModeEnabled() bool {
	return activeFrontendMode == frontendModeCatalogOnly
}

func cartFeatureEnabled() bool {
	return !catalogOnlyModeEnabled()
}

func assistantFeatureEnabled() bool {
	return assistantEnabled && !catalogOnlyModeEnabled()
}

func currentProductPrice(product *pb.Product) *pb.Money {
	if product != nil && product.GetPriceUsd() != nil {
		return product.GetPriceUsd()
	}

	return &pb.Money{CurrencyCode: defaultCurrency}
}

func (fe *frontendServer) catalogOnlyUnavailableHandler(feature string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
		renderHTTPError(log, r, w, errors.Errorf("%s is disabled when %s=%s", feature, frontendModeEnvVar, frontendModeCatalogOnly), http.StatusNotFound)
	}
}
