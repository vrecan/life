package life

import (
	"sync"
)

//Life handles the creation of the background thread and shutdown management
type Life struct {
	wg   *sync.WaitGroup
	Done chan struct{}
	run  func()
	once *sync.Once
}

//NewLife creates life with the expected defaults
func NewLife() *Life {
	return &Life{wg: &sync.WaitGroup{}, Done: make(chan struct{}, 10), once: &sync.Once{}}
}

//Start the background thread.
func (l Life) Start() {
	l.once.Do(func() {
		l.wg.Add(1)
		go l.run()
	})

}

//SetRun will set the run function that will be called by Start.
func (l *Life) SetRun(f func()) {
	l.run = f
}

//WGAdd will add to lifes waitgroup.
func (l Life) WGAdd(i int) {
	l.wg.Add(i)
}

//WGDone will decrement lifes waitgroup.
func (l Life) WGDone() {
	l.wg.Done()
}

//Close will wait for he background thread to finsh and then exit
func (l Life) Close() error {
	l.Done <- struct{}{}
	l.wg.Wait()
	return nil
}
