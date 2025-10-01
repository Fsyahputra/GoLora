package main

import (
	"fmt"
	"time"

	"github.com/Fsyahputra/GoLora/Lora/SX1276"
	"github.com/Fsyahputra/GoLora/driver/periphIO"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

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

func getSpiConf(mod int) (*periphIO.SpiConf, string, string) {
	defConf := periphIO.NewDefaultConf()
	if mod == 1 {
		defConf.Freq = 10 * physic.MegaHertz
		return defConf, "GPIO36", "GPIO133"
	} else if mod == 0 {
		defConf.Freq = 10 * physic.MegaHertz
		defConf.Reg = "/dev/spidev4.0"
		return defConf, "GPIO38", "GPIO134"
	}
	return defConf, "", ""
}

func main() {
	_, err := host.Init()
	if err != nil {
		panic(err)
	}
	spiConf1, rst1, cb1 := getSpiConf(0)
	drv1, err := periphIO.NewDriver(cb1, rst1, spiConf1)
	if err != nil {
		panic(err)
	}
	hwDrv1, err := drv1.Init()
	if err != nil {
		panic(err)
	}

	gl := SX1276.NewGoLoraSX1276(hwDrv1, *NewMinimalLoraConf())
	gl.Begin()

	_, err = gl.RegisterCb(SX1276.OnTxDone, func() {
		fmt.Println("Pesan Terkirim")
	})
	defer func() {
		gl.Destroy()
	}()
	if err != nil {
		return
	}
	//
	time.Sleep(1 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	stopTimer := time.After(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			fmt.Println("hello ")
			err := gl.SendPacketWithTxCb([]byte("Halo Semua"))
			if err != nil {
				fmt.Println(err.Error())
			}
		case <-stopTimer:
			return
		}

	}
	fmt.Println("pesan sudah terkirim")
	time.Sleep(100 * time.Second)
}
