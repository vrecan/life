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
	v := NewSingleLife()

	started := waitOnChan(v.started, 5*time.Millisecond)
	if started == nil {
		t.Fatalf("SingleLife started when it wasn't supposed to")
	}
	terminated := waitOnChan(v.terminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("SingleLife terminated when it wasn't supposed to")
	}

	v.Start()
	started = waitOnChan(v.started, 5*time.Millisecond)
	ok(t, started)

	terminated = waitOnChan(v.terminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("SingleLife terminated when it wasn't supposed to")
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
	p := NewLifeWithChildren()

	started := waitOnChan(p.started, 5*time.Millisecond)
	if started == nil {
		t.Fatalf("LifeWithChildren started when it wasn't supposed to")
	}
	terminated := waitOnChan(p.terminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("LifeWithChildren terminated when it wasn't supposed to")
	}
	if len(p.childrenStarted) > 0 {
		t.Fatalf("Subroutines have started when they weren't supposed to")
	}
	if len(p.childrenTerminated) > 0 {
		t.Fatalf("Subroutines have started when they weren't supposed to")
	}

	// Start LifeWithChildren and make sure that both the main goroutine and its subroutines are running
	p.Start()
	started = waitOnChan(p.started, 5*time.Millisecond)
	ok(t, started)

	terminated = waitOnChan(p.terminated, 5*time.Millisecond)
	if terminated == nil {
		t.Fatalf("LifeWithChildren terminated when it wasn't supposed to")
	}

	for i := 0; i < p.numChildren; i++ {
		err := waitOnChan(p.childrenStarted, 5*time.Millisecond)
		ok(t, err)
	}
	if len(p.childrenStarted) > 0 {
		t.Fatalf("Too many subroutines started")
	}

	terminated = waitOnChan(p.childrenTerminated, 5*time.Millisecond)
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
	for i := 0; i < p.numChildren; i++ {
		err := waitOnChan(p.childrenTerminated, 5*time.Millisecond)
		ok(t, err)
	}
	if len(p.childrenTerminated) > 0 {
		t.Fatalf("Too many subroutines terminated")
	}
	if len(p.childrenStarted) > 0 {
		t.Fatalf("Subroutine has started when it shouldn't have")
	}
}

type SingleLife struct {
	*Life

	started    chan struct{}
	terminated chan struct{}
}

func NewSingleLife() SingleLife {
	l := SingleLife{
		Life:       NewLife(),
		started:    make(chan struct{}, 0),
		terminated: make(chan struct{}, 0),
	}
	l.SetRun(l.run)
	return l
}

func (v SingleLife) run() {
	close(v.started)
	select {
	case <-v.Life.Done:
		// Sleep to make sure that life waits for this to finish rather than returning immediately
		time.Sleep(5 * time.Millisecond)
		close(v.terminated)
	}
}

type LifeWithChildren struct {
	*Life

	started    chan struct{}
	terminated chan struct{}

	numChildren int

	childrenStarted    chan struct{}
	childrenTerminated chan struct{}
}

func NewLifeWithChildren() LifeWithChildren {
	numSubRoutines := 5
	p := LifeWithChildren{
		Life:               NewLife(),
		started:            make(chan struct{}, 0),
		terminated:         make(chan struct{}, 0),
		numChildren:        numSubRoutines,
		childrenStarted:    make(chan struct{}, numSubRoutines),
		childrenTerminated: make(chan struct{}, numSubRoutines),
	}
	p.SetRun(p.run)
	return p
}

func (p LifeWithChildren) run() {
	defer close(p.terminated)
	close(p.started)
	for i := 0; i < p.numChildren; i++ {
		p.Life.WGAdd(1)
		go p.subRoutine()
	}
	select {
	case <-p.Life.Done:
		return
	}
}

func (p LifeWithChildren) subRoutine() {
	defer p.Life.WGDone()
	p.childrenStarted <- struct{}{}

	select {
	case <-p.Life.Done:
		// Same as above: make sure that life waits for this to finish
		time.Sleep(5 * time.Millisecond)
		p.childrenTerminated <- struct{}{}
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
