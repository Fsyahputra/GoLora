package main

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

type SoftPwm struct {
	Pin       gpio.PinIO
	DutyCycle int
	Frequency physic.Frequency
	dhch      chan int
}

func (p *SoftPwm) Start() chan struct{} {
	p.dhch = make(chan int)
	stpCh := make(chan struct{})
	period := time.Second / time.Duration(p.Frequency)
	go func() {
		for {
			select {
			case dc := <-p.dhch:
				p.DutyCycle = dc
			case <-stpCh:
				break
			default:

			}
			highDuration := time.Duration(float64(period) * (float64(p.DutyCycle) / 100.0))
			lowDuration := period - highDuration
			p.Pin.Out(gpio.High)
			time.Sleep(highDuration)
			p.Pin.Out(gpio.Low)
			time.Sleep(lowDuration)
		}
	}()
	return stpCh
}

func main() {
	_, err := host.Init()
	if err != nil {
		panic(err)
	}
	p := gpioreg.ByName("GPIO134")
	if p == nil {
		panic("Failed to find GPIO134")
	}
	pwm := &SoftPwm{
		Pin:       p,
		DutyCycle: 0,
		Frequency: 200 * physic.Hertz,
	}
	i := 0
	_ = pwm.Start()
	isHigh := false
	for {
		if i >= 100 {
			isHigh = true
		}

		if i <= 0 {
			isHigh = false
		}

		if isHigh {
			i -= 1
		}

		if !isHigh {
			i += 1
		}
		pwm.dhch <- i
		time.Sleep(50 * time.Millisecond)
		fmt.Println(i)
	}
	//close(stopCh)
}
