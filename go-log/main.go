package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type Application struct {
	logInfo  *log.Logger
	logDebug *log.Logger
}

type option func(*Application) option

func (a *Application) Option(opts ...option) (undo []option) {
	for _, opt := range opts {
		prev := opt(a)
		undo = append(undo, prev)
	}
	return
}

func DebugOff() option {
	return debugTo(io.Discard)
}

func debugTo(w io.Writer) option {
	return func(a *Application) option {
		prev := a.logDebug.Writer()
		a.logDebug.SetOutput(w)
		return debugTo(prev)

	}
}

func NewApplication(logTo io.Writer) *Application {
	app := &Application{}
	lgFlags := log.Ldate | log.Ltime | log.Lshortfile | log.Lmsgprefix

	app.logInfo = log.New(logTo, "[INFO] ", lgFlags)
	app.logDebug = log.New(logTo, "[DEBUG] ", lgFlags)

	return app
}

func (a *Application) infof(name string) func(string, ...any) {
	return func(format string, v ...any) {
		a.logInfo.Printf(name+" "+format, v...)
	}
}

func (a *Application) debugf(name string) func(string, ...any) {
	return func(format string, v ...any) {
		a.logDebug.Printf(name+" "+format, v...)
	}
}

func (a *Application) discard() func(string, ...any) {
	return func(s string, a ...any) {}
}

func (a *Application) Run(done context.Context) {

	infof := a.infof("Run")
	// example of discarding logging for a specific function
	debugf := a.discard()

	infof("start")
	defer infof("end")

	var wg sync.WaitGroup
	wg.Add(2)

	debugf("launch first task")
	go func() {
		defer wg.Done()
		a.task(done, "Task 1")
	}()

	debugf("launch second task")
	go func() {
		defer wg.Done()
		a.task(done, "Task 2")
	}()

	wg.Wait()
}

func (a *Application) task(done context.Context, name string) {

	infof := a.infof(name)
	debugf := a.debugf(name)

	infof("start")
	defer infof("end")

	tkr := time.NewTicker(400 * time.Millisecond)
	defer tkr.Stop()

	for {
		select {
		case <-done.Done():
			return
		case <-tkr.C:
			debugf("still running")
		}
	}
}

var dbg = flag.Bool("v", false, "verbose output. enable debug")

func main() {

	flag.Parse()

	myapp := NewApplication(os.Stdout)

	if !*dbg {
		myapp.Option(DebugOff())
	}

	done, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	myapp.Run(done)
}
