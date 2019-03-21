package life

import (
	"sync"
	"sync/atomic"
)

var (
	open   int32 = 0
	closed int32 = 1
)

// Life handles the creation of the background thread and shutdown management
type Life struct {
	wg     *sync.WaitGroup
	Done   chan struct{}
	run    func()
	once   *sync.Once
	closed *int32
}

// NewLife creates life with the expected defaults
func NewLife() *Life {
	return &Life{
		wg:     &sync.WaitGroup{},
		Done:   make(chan struct{}, 0),
		once:   &sync.Once{},
		closed: intptr(open),
	}
}

func intptr(i int32) *int32 {
	return &i
}

// Start the background thread.
func (l Life) Start() {
	l.once.Do(func() {
		l.WGAdd(1)
		go l.runner()
	})
}

func (l Life) runner() {
	defer l.wg.Done()
	l.run()
}

// SetRun will set the run function that will be called by Start.
func (l *Life) SetRun(f func()) {
	l.run = f
}

// WGAdd will add to life's waitgroup.
func (l Life) WGAdd(i int) {
	l.wg.Add(i)
}

// WGDone will decrement life's waitgroup.
func (l Life) WGDone() {
	l.wg.Done()
}

// Close will wait for the background thread to finish and then exit
func (l Life) Close() error {
	closeNow := atomic.CompareAndSwapInt32(l.closed, open, closed)
	if closeNow {
		close(l.Done)
	}
	l.wg.Wait()
	return nil
}
