package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Fsyahputra/GoLora/Lora/SX1276"
	"github.com/Fsyahputra/GoLora/driver"
	"github.com/Fsyahputra/GoLora/driver/periphIO"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

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

func Mod1Daemon(drv *driver.Driver, wg *sync.WaitGroup) {
	defer wg.Done()
	gl := SX1276.NewGoLoraSX1276(drv, *NewMinimalLoraConf())
	err := gl.Begin()
	if err != nil {
		log.Fatal(err)
		return
	}
	err = gl.CheckConn()
	if err != nil {
		log.Fatal(err)
		return
	}
	// registers, err := gl.DumpRegisters()
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	//
	// for addr, val := range registers {
	// 	log.Printf("gl mod 1 Reg 0x%02X: 0x%02X\n", addr, val)
	// }
	fmt.Println("waiting data")
	cb, err := gl.RegisterCb(SX1276.OnRxDone, func() {
		data, err := gl.ReceivePacket()
		if err != nil {
			log.Println("Error reading packet:", err)
			return
		}
		fmt.Println("data received from other lora", string(data))
	})
	if err != nil {
		return
	}
	time.Sleep(360 * time.Minute)
	close(cb)
	err = gl.Destroy()
	if err != nil {
		return
	}

}

func Mod0Daemon(drv *driver.Driver, wg *sync.WaitGroup) {
	defer wg.Done()
	gl0 := SX1276.NewGoLoraSX1276(drv, *NewMinimalLoraConf())
	err := gl0.Begin()
	if err != nil {
		log.Fatal(err)
		return
	}
	err = gl0.CheckConn()
	if err != nil {
		log.Fatal(err)
		return
	}

	registers, err := gl0.DumpRegisters()
	if err != nil {
		log.Fatal(err)
		return
	}

	for addr, val := range registers {
		log.Printf("gl0 mod 0 Reg 0x%02X: 0x%02X\n", addr, val)
	}
	fmt.Println("Switching to TX mode on mod 0")
	gl0.SendPacket([]byte("Hello from mod 0"))
	fmt.Println("Switching to RX mode on mod 1")
	gl0.SendPacket([]byte("Hello from mod 1"))
	// ticker := time.NewTicker(1000 * time.Millisecond)
	// i := 1
	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		fmt.Println("hello")
	// 		if err := gl0.SendPacket([]byte(fmt.Sprintf("Hello from mod 1 packet %d", i))); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	}
	// 	i++
	// }
}

func main() {
	var wg sync.WaitGroup
	_, err := host.Init()
	if err != nil {
		panic(err)
	}
	// spiConf0, rst0, cb0 := getSpiConf(0)
	spiConf1, rst1, cb1 := getSpiConf(1)

	// drv0, err := periphIO.NewDriver(cb0, rst0, spiConf0)
	// if err != nil {
	// 	panic(err)
	// }
	drv1, err := periphIO.NewDriver(cb1, rst1, spiConf1)
	if err != nil {
		panic(err)
	}
	// hwDrv0, err := drv0.Init()
	// if err != nil {
	// 	panic(err)
	// }
	hwDrv1, err := drv1.Init()
	if err != nil {
		panic(err)
	}
	wg.Add(2)
	// go Mod0Daemon(hwDrv0, &wg)
	go Mod1Daemon(hwDrv1, &wg)
	wg.Wait()
}
