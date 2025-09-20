package main

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

func readVal() {
	reader := gpioreg.ByName("GPIO134")
	if reader == nil {
		panic("Failed to find GPIO134")
	}
	if err := reader.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		panic(err)
	}
	fmt.Println("reading GPIO134")
	//val := reader.Read()
	for {
		fmt.Printf("Value changed to %v\n", reader.Read())
		time.Sleep(500 * time.Millisecond)
	}

}

func main() {
	_, err := host.Init()
	if err != nil {
		panic(err)
	}
	p := gpioreg.ByName("GPIO133")
	go readVal()
	for {
		p.Out(gpio.High)
		time.Sleep(1 * time.Second)
		p.Out(gpio.Low)
		time.Sleep(1 * time.Second)
	}
}
