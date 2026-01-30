package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Fsyahputra/GoLora/Lora/SX1276"
	"github.com/Fsyahputra/GoLora/driver/periphIO"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

func main() {
	_, err := host.Init()
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: program <payload>")
	}
	ctx := context.Background()

	payload := []byte(os.Args[1])

	spiConf0, rst0, cb0 := getSpiConf(0)

	drv0, err := periphIO.NewDriver(cb0, rst0, spiConf0)
	if err != nil {
		panic(err)
	}

	hwDrv0, err := drv0.Init()
	if err != nil {
		panic(err)
	}

	gl0 := SX1276.NewGoLoraSX1276(hwDrv0, *NewMinimalLoraConf())

	if err := gl0.Begin(); err != nil {
		log.Fatal(err)
	}

	if err := gl0.CheckConn(); err != nil {
		log.Fatal(err)
	}

	// Kirim payload sekali
	fmt.Printf("Sending packet: %s\n", payload)
	if err := gl0.SendPacket(ctx, payload); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Packet sent successfully!")
}

func getSpiConf(mod int) (*periphIO.SpiConf, string, string) {
	defConf := periphIO.NewDefaultConf()
	if mod == 0 {
		defConf.Freq = 10 * physic.MegaHertz
		return defConf, "GPIO135", "GPIO134"
	} else if mod == 1 {
		defConf.Freq = 10 * physic.MegaHertz
		defConf.Reg = "/dev/spidev4.0"
		return defConf, "GPIO38", "GPIO63"
	}
	return defConf, "", ""
}

func NewMinimalLoraConf() *SX1276.LoraConf {
	return &SX1276.LoraConf{
		TxPower:        17,
		SF:             7,
		BW:             125000,
		Denum:          1,
		PreambleLength: 8,
		SyncWord:       0x34,
		Frequency:      868000001,
		Header:         true,
		EnableCrc:      true,
	}
}
