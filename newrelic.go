// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
)

var (
	app     *newrelic.Application
	wrapper = newrelic.WrapHandle
)

func init() {
	log.SetPrefix("main")
	name := os.Getenv("NEWRELIC_NAME")
	lice := os.Getenv("NEWRELIC_LICENSE")

	if name == "" || lice == "" {
		// Don't use NewRelic is name or license is missing.
		wrapper = func(app *newrelic.Application, pattern string, handler http.Handler) (string, http.Handler) {
			return pattern, handler
		}
		log.Println("NewRelic is deactivated.")
		return
	}

	var err error
	app, err = newrelic.NewApplication(
		newrelic.ConfigAppName(os.Getenv("NEWRELIC_NAME")),
		newrelic.ConfigLicense(os.Getenv("NEWRELIC_LICENSE")),
		newrelic.ConfigDistributedTracerEnabled(true),
	)
	if err != nil {
		log.Fatalf("Failed to created NewRelic application: %v", err)
	}

	log.Println("NewRelic is activated.")
}
