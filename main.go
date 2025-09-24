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
		return defConf, "GPIO36", "GPIO133"

	} else if mod == 1 {
		defConf.Freq = 10 * physic.MegaHertz

		defConf.Reg = "/dev/spidev4.0"
		return defConf, "GPIO38", "GPIO134"
	}
	return defConf, "", ""
}

func NewMinimalLoraConf() *SX1276.LoraConf {
	return &SX1276.LoraConf{
		TxPower:        14,        // dayanya cukup standar, 14 dBm
		SF:             7,         // Spreading Factor minimal untuk komunikasi standar
		BW:             125000,    // Bandwidth 125 kHz (default LoRa)
		Denum:          1,         // Coding rate 4/5 (1 = 4/5)
		PreambleLength: 8,         // minimal preamble
		SyncWord:       0x34,      // sync word standar LoRaWAN
		Frequency:      868000000, // Frekuensi default 868 MHz (ubah sesuai region)
		Header:         true,      // explicit header
		EnableCrc:      true,      // CRC aktif
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
	registers, err := gl.DumpRegisters()
	if err != nil {
		log.Fatal(err)
		return
	}

	for addr, val := range registers {
		log.Printf("gl mod 1 Reg 0x%02X: 0x%02X\n", addr, val)
	}
	//_, err = gl.RegisterCb(SX1276.OnRxDone, func() {
	//	fmt.Println("Packet received on mod 1")
	//	data, err := gl.ReceivePacket()
	//	if err != nil {
	//		log.Println("Error reading packet:", err)
	//		return
	//	}
	//	fmt.Printf("Received data on mod 1: %s\n", string(data))
	//
	//})
	for {
		gl.ChangeMode(SX1276.RxContinuous)
		time.Sleep(1 * time.Second)
		data, err := gl.ReceivePacket()
		if err != nil {
			log.Println("Error reading packet:", err)
			return
		}
		fmt.Printf("Received data on mod 1: %s\n", string(data))
		if err != nil {
			log.Fatal(err)
			return
		}
		//gl.Begin()
		gl.ChangeMode(SX1276.RxContinuous)
		time.Sleep(1 * time.Second)
		data, err = gl.ReceivePacket()
		if err != nil {
			log.Println("Error reading packet:", err)
			return
		}
		fmt.Printf("Received data on mod 1: %s\n", string(data))
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	return

}

func Mod0Daemon(drv *driver.Driver, wg *sync.WaitGroup) {
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

	registers, err := gl.DumpRegisters()
	if err != nil {
		log.Fatal(err)
		return
	}

	for addr, val := range registers {
		log.Printf("gl mod 0 Reg 0x%02X: 0x%02X\n", addr, val)
	}

	gl.ChangeMode(SX1276.Tx)
	//gl.SendPacket([]byte("Hello from mod 0asdasdasd"))

	ticker := time.NewTicker(100 * time.Millisecond)
	timeout := time.After(100 * time.Second) // cukup sekali di-declare

	i := 1
	for {
		select {
		case <-ticker.C:
			if err := gl.SendPacket([]byte(fmt.Sprintf("hello from mod 0 n packet %d", i))); err != nil {
				log.Fatal(err)
			}
			i++
		case <-timeout:
			//ticker.Stop()
			//if err := gl.Destroy(); err != nil {
			//	log.Fatal(err)
			//}
			return
		}
	}
	time.Sleep(3 * time.Second)
	//gl.SendPacket([]byte("Goodbye from mod 0"))

}

func main() {
	var wg sync.WaitGroup
	_, err := host.Init()
	if err != nil {
		panic(err)
	}
	spiConf0, rst0, cb0 := getSpiConf(0)
	spiConf1, rst1, cb1 := getSpiConf(1)

	drv0, err := periphIO.NewDriver(cb0, rst0, spiConf0)
	if err != nil {
		panic(err)
	}
	drv1, err := periphIO.NewDriver(cb1, rst1, spiConf1)
	if err != nil {
		panic(err)
	}
	hwDrv0, err := drv0.Init()
	if err != nil {
		panic(err)
	}
	hwDrv1, err := drv1.Init()
	if err != nil {
		panic(err)
	}
	wg.Add(2)
	go Mod0Daemon(hwDrv0, &wg)
	go Mod1Daemon(hwDrv1, &wg)
	wg.Wait()
	fmt.Println("All done")
}
