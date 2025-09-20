package periphIO

import (
	"errors"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type RSTPin struct {
	pin gpio.PinIO
}

func NewRstPinPeriphIO(pinName string) (*RSTPin, error) {
	p := gpioreg.ByName(pinName)
	if p == nil {
		return nil, errors.New("RstPin GPIO does not exist")
	}
	rst := &RSTPin{pin: p}
	return rst, nil
}

func (r *RSTPin) toggle(level gpio.Level) error {
	if err := r.pin.Out(level); err != nil {
		return err
	}
	return nil
}
func (r *RSTPin) Low() error {
	if err := r.toggle(gpio.Low); err != nil {
		return err
	}
	return nil
}

func (r *RSTPin) High() error {
	if err := r.toggle(gpio.High); err != nil {
		return err
	}
	return nil
}
