package life

import (
	"sync"
)

type Life struct {
	wg   *sync.WaitGroup
	done chan struct{}
	run  func()
	once *sync.Once
}

func NewLife() *Life {
	return &Life{wg: &sync.WaitGroup{}, done: make(chan struct{}, 10), once: &sync.Once{}}
}

func (l Life) Start() {
	l.once.Do(func() {
		l.wg.Add(1)
		go l.run()
	})

}

func (l *Life) SetRun(f func()) {
	l.run = f
}

func (l Life) WGAdd(i int) {
	l.wg.Add(i)
}

func (l Life) WGDone() {
	l.wg.Done()
}

//Close will wait for he background thread to finsh and then exit
func (l Life) Close() error {
	l.done <- struct{}{}
	l.wg.Wait()
	return nil
}
