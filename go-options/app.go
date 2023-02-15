package main

import (
	"fmt"
	"time"
)

// App is a test application
type App struct {
	Name      string
	Verbosity int
	MaxAlive  time.Duration
}

func (a *App) String() string {
	return fmt.Sprintf("%s: verbosity [%d], max alive [%v]", a.Name, a.Verbosity, a.MaxAlive)
}

func (a *App) Run() {
	undo := a.Option(Verbosity(2), MaxAlive(1*time.Second))
	defer a.Option(undo...)

	fmt.Println("RUN:", a)

	defer func() func() {
		start := time.Now()
		return func() {
			fmt.Printf("Done in: %v\n", time.Since(start).Round(a.MaxAlive))
		}
	}()()

	time.Sleep(a.MaxAlive)
}
