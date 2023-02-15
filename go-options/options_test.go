package main

import (
	"testing"
)

func TestVerbosity(t *testing.T) {
	app := &App{Name: "test app"}

	undo := app.Option(Verbosity(4))
	if app.Verbosity != 4 {
		t.Fatal("Verbosity not set")
	}
	undo[0](app)
	if app.Verbosity != 0 {
		t.Fatal("Verbosity not properly restored")
	}
}
