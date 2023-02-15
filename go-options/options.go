package main

import "time"

type option func(*App) option

// Option sets options to the passed in values.
// It returns a list of functions that can be called to restore the options
// to their previous values
func (a *App) Option(opts ...option) (undo []option) {

	for _, opt := range opts {
		prev := opt(a)
		undo = append(undo, prev)
	}

	return
}

// Verbosity sets the logs verbosity.
// It returns a function that can be used to restore the verbosity to its previous value
func Verbosity(v int) option {
	return func(a *App) option {
		prev := a.Verbosity
		a.Verbosity = v
		return Verbosity(prev)
	}
}

// MaxAlive sets the maximum time the application can run.
// It returns a function that can be used to restore max alive to its previous value
func MaxAlive(d time.Duration) option {
	return func(a *App) option {
		prev := a.MaxAlive
		a.MaxAlive = d
		return MaxAlive(prev)
	}
}
