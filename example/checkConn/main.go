package main

import (
	"fmt"
	// "time"

	"github.com/Fsyahputra/GoLora/Lora/SX1276"
	"github.com/Fsyahputra/GoLora/driver/periphIO"
	"periph.io/x/host/v3"
)

func NewMinimalLoraConf() *SX1276.LoraConf {
	return &SX1276.LoraConf{
		TxPower:        17,
		SF:             7,
		BW:             125000,
		Denum:          5,
		PreambleLength: 8,
		SyncWord:       0x12,
		Frequency:      915000000,
		Header:         true,
		EnableCrc:      true,
	}
}

func main() {
	if _, err := host.Init(); err != nil {
		panic(err)
	}
	cbpin2, rstpin2 := "GPIO134", "GPIO135" // ujung  gnd bawah, atas1, atas2
	spiconf2 := periphIO.NewDefaultConf()

	drv2, err := periphIO.NewDriver(cbpin2, rstpin2, spiconf2)
	if err != nil {
		panic(err)
	}
	hwDrv2, err := drv2.Init()
	if err != nil {
		panic(err)
	}
	loraConf2 := NewMinimalLoraConf()
	lora2 := SX1276.NewGoLoraSX1276(hwDrv2, *loraConf2)
	if err := lora2.Begin(); err != nil {
		panic(err)
	}

	fmt.Println("connected to lora device 2")
	// regVal, err := lora2.DumpRegisters()
	// for _, reg := range regVal {
	// 	fmt.Printf("0x%02X, 0x%02X\n", reg.Reg, reg.Val)
	// }

	cbPin1, rstPin1 := "GPIO63", "GPIO38"
	spiconf1 := periphIO.NewDefaultConf()
	spiconf1.Reg = "/dev/spidev4.0"
	drv1, err := periphIO.NewDriver(cbPin1, rstPin1, spiconf1)
	if err != nil {
		panic(err)
	}
	hwDrv1, err := drv1.Init()
	if err != nil {
		panic(err)
	}
	loraConf1 := NewMinimalLoraConf()
	lora1 := SX1276.NewGoLoraSX1276(hwDrv1, *loraConf1)
	if err := lora1.Begin(); err != nil {
		panic(err)
	}
	fmt.Println("connected to lora device 1")

}
