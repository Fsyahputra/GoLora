package RstPin

import (
	"errors"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type RSTPinPeriphIO struct {
	pin gpio.PinIO
}

func NewRstPinPeriphIO(pinName string) (*RSTPinPeriphIO, error) {
	p := gpioreg.ByName(pinName)
	if p == nil {
		return nil, errors.New("GPIO does not exist")
	}
	rst := &RSTPinPeriphIO{pin: p}
	return rst, nil
}

func (r *RSTPinPeriphIO) toggle(level gpio.Level) error {
	if err := r.pin.Out(level); err != nil {
		return err
	}
	return nil
}
func (r *RSTPinPeriphIO) Low() error {
	if err := r.toggle(gpio.Low); err != nil {
		return err
	}
	return nil
}

func (r *RSTPinPeriphIO) High() error {
	if err := r.toggle(gpio.High); err != nil {
		return err
	}
	return nil
}
