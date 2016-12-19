package life

import (
	"testing"

	log "github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGivingLife(t *testing.T) {
	defer log.Flush()

	Convey("Give vrecan life", t, func() {
		vrecan := NewVrecan()
		defer vrecan.Close()
		vrecan.Start()

	})
	Convey("increment and decrement count from outside", t, func() {
		vrecan := NewVrecan()
		defer vrecan.Close()
		vrecan.Start()
		vrecan.WGAdd(1)
		vrecan.WGDone()
	})
}

type Vrecan struct {
	*Life
}

func NewVrecan() *Vrecan {
	vrecan := &Vrecan{}
	vrecan.Life = NewLife()
	vrecan.SetRun(vrecan.run)
	return vrecan
}

func (v Vrecan) run() {
Main:
	for {
		select {
		case <-v.Done:
			log.Debug("Done chan")
			break Main
		}
	}
}
