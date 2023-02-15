package main

import (
	"fmt"
	"time"
)

func Example() {

	app := &App{Name: "test app"}

	app.Option(Verbosity(3), MaxAlive(4*time.Second))
	fmt.Println(app)
	app.Run()
	fmt.Println(app)
	// Output:
	// test app: verbosity [3], max alive [4s]
	// RUN: test app: verbosity [2], max alive [1s]
	// Done in: 1s
	// test app: verbosity [3], max alive [4s]
}
