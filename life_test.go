package life

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestLife(t *testing.T) {
	v := NewVrecan()

	started := waitOnChan(v.started, 5*time.Millisecond)
	if started == nil {
		t.Fatalf("Vrecan started when it wasn't supposed to")
	}
	terminated := waitOnChan(v.terminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("Vrecan terminated when it wasn't supposed to")
	}

	v.Start()
	started = waitOnChan(v.started, 5*time.Millisecond)
	ok(t, started)

	terminated = waitOnChan(v.terminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("Vrecan terminated when it wasn't supposed to")
	}

	// Set up a maximum wait time before failing
	timer := time.NewTimer(50 * time.Millisecond)
	defer timer.Stop()

	errChan := make(chan error, 1)
	go func() {
		errChan <- v.Close()
	}()

	select {
	case <-timer.C:
		t.Fatalf("Timed out waiting for close to finish")
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Error received from Close call: %s", err)
		}
	}

	terminated = waitOnChan(v.terminated, 5*time.Millisecond)
	ok(t, terminated)
}

func TestLife_multiRoutine(t *testing.T) {
	// TODO rename variable
	p := NewPcman312()

	started := waitOnChan(p.started, 5*time.Millisecond)
	if started == nil {
		t.Fatalf("pcman312 started when it wasn't supposed to")
	}
	terminated := waitOnChan(p.terminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("pcman312 terminated when it wasn't supposed to")
	}
	if len(p.subRoutinesStarted) > 0 {
		t.Fatalf("Subroutines have started when they weren't supposed to")
	}
	if len(p.subRoutinesTerminated) > 0 {
		t.Fatalf("Subroutines have started when they weren't supposed to")
	}

	// Start pcman312 and make sure that both the main goroutine and its subroutines are running
	p.Start()
	started = waitOnChan(p.started, 5*time.Millisecond)
	ok(t, started)

	terminated = waitOnChan(p.terminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("pcman312 terminated when it wasn't supposed to")
	}

	for i := 0; i < p.numSubRoutines; i++ {
		err := waitOnChan(p.subRoutinesStarted, 5*time.Millisecond)
		ok(t, err)
	}
	if len(p.subRoutinesStarted) > 0 {
		t.Fatalf("Too many subroutines started")
	}

	terminated = waitOnChan(p.subRoutinesTerminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("Subroutines terminated when they weren't supposed to")
	}

	// Set up a maximum wait time before failing
	timer := time.NewTimer(50 * time.Millisecond)
	defer timer.Stop()

	errChan := make(chan error, 1)
	go func() {
		errChan <- p.Close()
	}()

	select {
	case <-timer.C:
		t.Fatalf("Timed out waiting for close to finish")
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Error received from Close call: %s", err)
		}
	}

	terminated = waitOnChan(p.terminated, 5*time.Millisecond)
	ok(t, terminated)

	// Check on the subroutines
	for i := 0; i < p.numSubRoutines; i++ {
		err := waitOnChan(p.subRoutinesTerminated, 5*time.Millisecond)
		ok(t, err)
	}
	if len(p.subRoutinesTerminated) > 0 {
		t.Fatalf("Too many subroutines terminated")
	}
	if len(p.subRoutinesStarted) > 0 {
		t.Fatalf("Subroutine has started when it shouldn't have")
	}
}

type Vrecan struct {
	*Life

	started    chan struct{}
	terminated chan struct{}
}

func NewVrecan() Vrecan {
	vrecan := Vrecan{
		Life:       NewLife(),
		started:    make(chan struct{}, 0),
		terminated: make(chan struct{}, 0),
	}
	vrecan.SetRun(vrecan.run)
	return vrecan
}

func (v Vrecan) run() {
	close(v.started)
	select {
	case <-v.Life.Done:
		// Sleep to make sure that life waits for this to finish rather than returning immediately
		time.Sleep(5 * time.Millisecond)
		close(v.terminated)
	}
}

type pcman312 struct {
	*Life

	started    chan struct{}
	terminated chan struct{}

	numSubRoutines int

	subRoutinesStarted    chan struct{}
	subRoutinesTerminated chan struct{}
}

func NewPcman312() pcman312 {
	numSubRoutines := 5
	p := pcman312{
		Life:                  NewLife(),
		started:               make(chan struct{}, 0),
		terminated:            make(chan struct{}, 0),
		numSubRoutines:        numSubRoutines,
		subRoutinesStarted:    make(chan struct{}, numSubRoutines),
		subRoutinesTerminated: make(chan struct{}, numSubRoutines),
	}
	p.SetRun(p.run)
	return p
}

func (p pcman312) run() {
	defer close(p.terminated)
	close(p.started)
	for i := 0; i < p.numSubRoutines; i++ {
		p.Life.WGAdd(1)
		go p.subRoutine()
	}
	select {
	case <-p.Life.Done:
		return
	}
}

func (p pcman312) subRoutine() {
	defer p.Life.WGDone()
	p.subRoutinesStarted <- struct{}{}

	select {
	case <-p.Life.Done:
		// Same as above: make sure that life waits for this to finish
		time.Sleep(5 * time.Millisecond)
		p.subRoutinesTerminated <- struct{}{}
	}
}

func waitOnChan(c chan struct{}, wait time.Duration) (err error) {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return errors.New("timed out")
	case <-c:
		return nil
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}
