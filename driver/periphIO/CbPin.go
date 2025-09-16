package periphIO

import (
	"errors"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type CbPin struct {
	Pin gpio.PinIn
}

func NewCbPin(pinName string) (*CbPin, error) {
	p := gpioreg.ByName(pinName)
	if p == nil {
		return nil, errors.New("GPIO does not exist")
	}
	pin := &CbPin{Pin: p}
	if err := pin.Init(); err != nil {
		return nil, err
	}
	return pin, nil
}

func (cbp *CbPin) Init() error {
	err := cbp.Pin.In(gpio.PullDown, gpio.RisingEdge)
	if err != nil {
		return err
	}
	return nil
}

func (cbp *CbPin) ReadVal() (bool, error) {
	value := cbp.Pin.Read()
	return bool(value), nil
}
