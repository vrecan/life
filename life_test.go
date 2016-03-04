package life

import (
	log "github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGivingLife(t *testing.T) {
	defer log.Flush()

	Convey("Give vrecan life", t, func() {
		vrecan := NewVrecan()
		defer vrecan.Close()
		vrecan.Start()

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
	defer v.Life.WGDone()
Main:
	for {
		select {
		case <-v.done:
			log.Debug("Done chan")
			break Main
		}
	}
}
