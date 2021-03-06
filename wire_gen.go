// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"
	"database/sql"
	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/handler"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opencensus.io/trace"
	"gocloud.dev/server"
	"gocloud.dev/server/driver"
	"gocloud.dev/server/health"
	"gocloud.dev/server/health/sqlhealth"
	"net/http"
)

// Injectors from inject_main.go:

func newServer(ctx context.Context, logger zerolog.Logger, dsn datastore.PGDatasourceName) (*server.Server, func(), error) {
	db, cleanup, err := datastore.NewDB(dsn, logger)
	if err != nil {
		return nil, nil, err
	}
	datastoreDatastore := datastore.NewDatastore(db)
	application := app.NewApplication(datastoreDatastore, logger)
	appHandler := handler.NewAppHandler(application)
	router := newRouter(appHandler)
	v, cleanup2 := appHealthChecks(db)
	exporter := _wireExporterValue
	sampler := trace.AlwaysSample()
	defaultDriver := server.NewDefaultDriver()
	options := &server.Options{
		HealthChecks:          v,
		TraceExporter:         exporter,
		DefaultSamplingPolicy: sampler,
		Driver:                defaultDriver,
	}
	serverServer := server.New(router, options)
	return serverServer, func() {
		cleanup2()
		cleanup()
	}, nil
}

var (
	_wireExporterValue = trace.Exporter(nil)
)

// inject_main.go:

// applicationSet is the Wire provider set for the application
var applicationSet = wire.NewSet(app.NewApplication, newRouter, wire.Bind(new(http.Handler), new(*mux.Router)), handler.NewAppHandler)

// goCloudServerSet
var goCloudServerSet = wire.NewSet(trace.AlwaysSample, server.New, server.NewDefaultDriver, wire.Bind(new(driver.Server), new(*server.DefaultDriver)))

// appHealthChecks returns a health check for the database. This will signal
// to Kubernetes or other orchestrators that the server should not receive
// traffic until the server is able to connect to its database.
func appHealthChecks(db *sql.DB) ([]health.Checker, func()) {
	dbCheck := sqlhealth.New(db)
	list := []health.Checker{dbCheck}
	return list, func() {
		dbCheck.Stop()
	}
}
