package SX1276

import (
	"errors"

	"github.com/Fsyahputra/GoLora/Lora/SX1276/internal"
)

type BitUtils interface {
	setWriteMask(reg byte) byte
	setReadMask(reg byte) byte
	changeMode(mode LoraMode) byte
	setTxPower(power byte) byte
	setFreq(freq uint64) []byte
	setSF(sf byte) byte
	setBW(bw byte) byte
	setCrc(enable bool, currentModemConfig byte) byte
	setHeader(header bool, currentModemConfig byte) byte
	setPreamble(length uint16) []byte
	checkData(irq byte) error
	setCodingRate(cr byte, currentModemConfig byte) byte
}

type LoraUtils struct{}

func (lu *LoraUtils) setCodingRate(cr byte, currentModemConfig byte) byte {
	clearedCrConf := currentModemConfig & 0xf1
	shiftedCr := cr << 1 & 0x0e
	overWrittenConf := clearedCrConf | shiftedCr
	return overWrittenConf
}

func (lu *LoraUtils) setHeader(header bool, currentModemConfig byte) byte {
	updatedConf := byte(0x00)
	if header {
		updatedConf = currentModemConfig & 0xfe
	} else {
		updatedConf = currentModemConfig | 0x01
	}
	return updatedConf
}

func (lu *LoraUtils) checkData(irq byte) error {
	if irq&internal.IRQ_RX_DONE_MASK == 0 {
		return errors.New("no Packet Received")
	}

	if irq&internal.IRQ_PAYLOAD_CRC_ERROR_MASK != 0 {
		return errors.New("packet damaged or lost in transmit")
	}
	return nil
}

func (lu *LoraUtils) setPreamble(length uint16) []byte {
	msbPreamble := byte(length>>8) & 0xff
	lsbPreamble := byte(length>>0) & 0xff
	return []byte{msbPreamble, lsbPreamble}
}

func (lu *LoraUtils) setCrc(enable bool, currentModemConfig byte) byte {
	crc := byte(0x00)
	updateConfig := currentModemConfig
	if enable {
		crc = crc | 0x04
		updateConfig = currentModemConfig | crc
	} else {
		crc = 0xfb
		updateConfig = currentModemConfig & crc
	}
	return updateConfig
}

func (lu *LoraUtils) setWriteMask(reg byte) byte {
	return reg | 0x80

}

func (lu *LoraUtils) setReadMask(reg byte) byte {
	return reg & 0x7f
}

func (lu *LoraUtils) changeMode(mode LoraMode) byte {
	loraMode := internal.MODE_LONG_RANGE_MODE
	var selectedMode byte = 0
	switch mode {
	case Sleep:
		selectedMode = loraMode | internal.MODE_SLEEP
	case Idle:
		selectedMode = loraMode | internal.MODE_STDBY
	case Tx:
		selectedMode = loraMode | internal.MODE_TX
	case RxContinuous:
		selectedMode = loraMode | internal.MODE_RX_CONTINUOUS
	case RxSingle:
		selectedMode = loraMode | internal.MODE_RX_SINGLE
	default:
		selectedMode = loraMode | internal.MODE_STDBY
	}
	return selectedMode
}

func (lu *LoraUtils) setTxPower(power byte) byte {
	return power | internal.PA_BOOST
}

func (lu *LoraUtils) setFreq(freq uint64) []byte {
	msb := byte(freq>>16) & 0xff
	mid := byte(freq>>8) & 0xff
	lsb := byte(freq) & 0xff
	return []byte{msb, mid, lsb}
}

func (lu *LoraUtils) setSF(sf byte) byte {
	sfReg := sf << 4 & 0xf0
	return sfReg
}

func (lu *LoraUtils) setBW(bw byte) byte {
	bwReg := bw << 4 & 0xf0
	return bwReg
}
