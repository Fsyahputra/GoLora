package SX1276

import (
	"time"

	"github.com/Fsyahputra/GoLora/Lora/SX1276/internal"
	"github.com/Fsyahputra/GoLora/driver"
)

type LoraConf struct {
	TxPower        uint8
	SF             uint8
	BW             uint8
	CodingRate     uint8
	PreambleLength uint8
	SyncWord       uint8
	Frequency      uint64
	ExplicitHeader bool
	EnableCrc      bool
}
type GoLora struct {
	driver.ModComm
	driver.RSTPin
	internal.LoraUtils
	Conf LoraConf
}

func NewGoLoraSX1276(modComm driver.ModComm, rstPin driver.RSTPin, conf LoraConf) *GoLora {
	gl := &GoLora{
		ModComm: modComm,
		RSTPin:  rstPin,
		Conf:    LoraConf{},
	}
	return gl
}

func (gl *GoLora) readReg(reg byte) (byte, error) {
	readReg := gl.SetReadMask(reg)
	rx, err := gl.ModComm.ReadFromMod(readReg)
	if err != nil {
		return 0, err
	}
	return rx, nil
}

func (gl *GoLora) writeReg(reg byte, value uint8) error {
	writeReg := gl.SetWriteMask(reg)
	byteValue := value
	if err := gl.ModComm.SendToMod(writeReg, byteValue); err != nil {
		return err
	}
	return nil
}

func (gl *GoLora) Reset() {
	gl.RSTPin.Low()
	time.Sleep(300 * time.Millisecond)
	gl.RSTPin.High()
	time.Sleep(1000 * time.Millisecond)
}

func (gl *GoLora) changeMode(mode internal.LoraMode) {
	modeVal := gl.LoraUtils.ChangeMode(mode)
	if err := gl.writeReg(internal.REG_OP_MODE, modeVal); err != nil {
		return
	}
}
