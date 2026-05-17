package periphIO

import (
	"errors"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type CbPin struct {
	Pin     gpio.PinIn
	pinName string
}

func NewCbPin(pinName string) (*CbPin, error) {
	p := gpioreg.ByName(pinName)
	if p == nil {
		return nil, errors.New("CbPin GPIO does not exist")
	}
	pin := &CbPin{Pin: p, pinName: pinName}
	if err := pin.Init(); err != nil {
		return nil, err
	}
	return pin, nil
}

func (cbp *CbPin) Init() error {
	err := cbp.Pin.In(gpio.PullNoChange, gpio.BothEdges)
	if err != nil {
		return err
	}

	// Apply Orange Pi H3 workaround for input stability
	applyGPIOInWorkaround(cbp.pinName)

	return nil
}

func (cbp *CbPin) ReadVal() (bool, error) {
	value := cbp.Pin.Read()
	return bool(value), nil
}
