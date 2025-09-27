package main

import (
	"fmt"
	"log"

	"github.com/Fsyahputra/GoLora/Lora/SX1276"
	"github.com/Fsyahputra/GoLora/driver/periphIO"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

func getSpiConf() (*periphIO.SpiConf, string, string) {
	defConf := periphIO.NewDefaultConf()
	defConf.Freq = 1 * physic.MegaHertz // 1 MHz
	return defConf, "GPIO38", "GPIO134"
}

func NewMinimalLoraConf() *SX1276.LoraConf {
	return &SX1276.LoraConf{
		TxPower:        14,
		SF:             7,
		BW:             125000,
		Denum:          1,
		PreambleLength: 8,
		SyncWord:       0x34,
		Frequency:      868000000,
		Header:         true,
		EnableCrc:      true,
	}
}

func main() {
	_, err := host.Init()
	if err != nil {
		panic(err)
	}

	spiConf, rst, cb := getSpiConf()
	drv, err := periphIO.NewDriver(cb, rst, spiConf)
	if err != nil {
		panic(err)
	}
	hwDrv, err := drv.Init()
	if err != nil {
		panic(err)
	}

	gl := SX1276.NewGoLoraSX1276(hwDrv, *NewMinimalLoraConf())
	if err := gl.Begin(); err != nil {
		log.Fatal(err)
	}
	if err := gl.CheckConn(); err != nil {
		log.Fatal(err)
	}

	cbHandle, err := gl.RegisterCb(SX1276.OnRxDone, func() {
		fmt.Println("Packet received")
		data, err := gl.ReceivePacket()
		if err != nil {
			log.Println("Error reading packet:", err)
			return
		}
		fmt.Printf("Received data: %s\n", string(data))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer close(cbHandle)

	// keep program running
	select {}
}
