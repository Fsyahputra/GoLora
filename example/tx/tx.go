package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Fsyahputra/GoLora/Lora/SX1276"
	"github.com/Fsyahputra/GoLora/driver/periphIO"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

func getSpiConf() (*periphIO.SpiConf, string, string) {
	defConf := periphIO.NewDefaultConf()
	defConf.Freq = 1 * physic.MegaHertz // 1 MHz
	return defConf, "GPIO36", "GPIO133"
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
	if len(os.Args) < 2 {
		fmt.Println("Usage: transmitter [single|burst]")
		return
	}
	mode := os.Args[1]

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

	gl.ChangeMode(SX1276.Tx)

	switch mode {
	case "single":
		if err := gl.SendPacket([]byte("Hello from mod 0")); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Packet sent (single mode)")

	case "burst":
		count := 10
		delayMs := 100
		for i := 0; i < count; i++ {
			if err := gl.SendPacket([]byte(fmt.Sprintf("Packet %d from mod 0", i+1))); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Packet %d sent\n", i+1)
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

	default:
		fmt.Println("Unknown mode. Use 'single' or 'burst'")
	}
}
