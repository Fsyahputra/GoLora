package SX1276

import (
	"errors"
	"fmt"
	"time"

	"github.com/Fsyahputra/GoLora/Lora/SX1276/internal"
	"github.com/Fsyahputra/GoLora/driver"
	"periph.io/x/conn/v3/physic"
)

type LoraConf struct {
	TxPower        uint8
	SF             uint8
	BW             uint8
	CodingRate     uint8
	PreambleLength uint8
	SyncWord       uint8
	Frequency      physic.Frequency
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
		Conf:    conf,
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

func (gl *GoLora) ChangeMode(mode internal.LoraMode) {
	modeVal := gl.LoraUtils.ChangeMode(mode)
	if err := gl.writeReg(internal.REG_OP_MODE, modeVal); err != nil {
		return
	}
}

func (gl *GoLora) SetTXPower(txPower uint8) {
	gl.Conf.TxPower = txPower
	tx := 0
	if txPower < 2 {
		fmt.Println("txPower Too Low set default tx=2")
		tx = 2
	} else if txPower > 17 {
		fmt.Println("txPower Too High set to 17")
		tx = 17
	}
	tx = tx - 2
	txReg := gl.LoraUtils.SetTxPower(byte(tx))
	if err := gl.writeReg(internal.REG_PA_CONFIG, txReg); err != nil {
		return
	}
}

func (gl *GoLora) writeRegMany(Regs []byte, Values []byte) error {
	if len(Regs) != len(Values) {
		return errors.New("len(Regs) != len(Values)")
	}
	type regValue struct {
		Reg   byte
		Value byte
	}
	regValues := make([]regValue, len(Regs))
	for i, Reg := range Regs {
		regValues[i] = regValue{Reg: Reg, Value: Values[i]}
	}
	for _, reg := range regValues {
		if err := gl.writeReg(reg.Reg, reg.Value); err != nil {
			return err
		}
	}

}

func (gl *GoLora) SetFrequency(freq physic.Frequency) {
	gl.Conf.Frequency = freq
	frf := (uint64(freq) << 19) / 32000000
	freqBytes := gl.LoraUtils.SetFreq(byte(frf))
	registers := []byte{internal.REG_FRF_MSB, internal.REG_FRF_MID, internal.REG_FRF_LSB}
	if err := gl.writeRegMany(registers, freqBytes); err != nil {
		return
	}
}

func (gl *GoLora) SetSF(sf uint8) {

	if sf < 6 {
		sf = 6
		fmt.Println("SF Too low set to 6")
	} else if sf > 12 {
		sf = 12
		fmt.Println("SF Too high set to 12")
	}
	gl.Conf.SF = sf
	sfReg := gl.LoraUtils.SetSF(sf)
	values := []byte{sfReg, internal.SF_DEF_OPTIMIZE, internal.SF_DEF_THRESHOLD}
	registers := []byte{internal.REG_MODEM_CONFIG_2, internal.REG_DETECTION_OPTIMIZE, internal.REG_DETECTION_THRESHOLD}
	defer func() {
		_ = gl.writeRegMany(registers, values)
	}()
	if sf == 6 {
		values[1] = internal.SF_6_OPTIMIZE
		values[2] = internal.SF_6_THRESHOLD
		return
	}
	return
}

func (gl *GoLora) SetBW(bw uint64) {
	var sbw uint8
	bwMaps := map[BW]uint8{
		BW_1: 1,
		BW_2: 2,
		BW_3: 3,
		BW_4: 4,
		BW_5: 5,
		BW_6: 6,
		BW_7: 7,
		BW_8: 8,
	}

	for thres, value := range bwMaps {
		if bw < uint64(thres) {
			sbw = value
		} else {
			sbw = 9
		}
	}
	gl.Conf.BW = sbw
	bwReg := gl.LoraUtils.SetBW(sbw)
	_ = gl.writeReg(internal.REG_MODEM_CONFIG_1, bwReg)
}

func (gl *GoLora) SetCrc(enable bool) {
	crcReg := gl.LoraUtils.SetCrc(enable)
	_ = gl.writeReg(internal.REG_MODEM_CONFIG_2, crcReg)
}

func (gl *GoLora) SetPreamble(length uint16) {

}
