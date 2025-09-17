package internal

import (
	"errors"
)

type LoraMode byte

const (
	Sleep        = LoraMode(MODE_SLEEP)
	Idle         = LoraMode(MODE_STDBY)
	Tx           = LoraMode(MODE_TX)
	RxSingle     = LoraMode(MODE_RX_SINGLE)
	RxContinuous = LoraMode(MODE_RX_CONTINUOUS)
)

type BitUtils interface {
	SetWriteMask(reg byte) byte
	SetReadMask(reg byte) byte
	ChangeMode(mode LoraMode) byte
	SetTxPower(power byte) byte
	SetFreq(freq uint64) []byte
	SetSF(sf byte) byte
	SetBW(bw byte) byte
	SetCrc(enable bool, currentModemConfig byte) byte
	SetHeader(header bool, currentModemConfig byte) byte
	SetPreamble(length uint16) []byte
	CheckData(irq byte) error
	SetCodingRate(cr byte, currentModemConfig byte) byte
}

type LoraUtils struct{}

func (lu *LoraUtils) SetCodingRate(cr byte, currentModemConfig byte) byte {
	clearedCrConf := currentModemConfig & 0xf1
	shiftedCr := cr << 1 & 0x0e
	overWrittenConf := clearedCrConf | shiftedCr
	return overWrittenConf
}

func (lu *LoraUtils) SetHeader(header bool, currentModemConfig byte) byte {
	updatedConf := byte(0x00)
	if header {
		updatedConf = currentModemConfig & 0xfe
	} else {
		updatedConf = currentModemConfig | 0x01
	}
	return updatedConf
}

func (lu *LoraUtils) CheckData(irq byte) error {
	var err error = nil
	if irq&IRQ_RX_DONE_MASK == 0 {
		err = errors.New("no Packet Received")
	}

	if irq&IRQ_PAYLOAD_CRC_ERROR_MASK == 0x20 {
		err = errors.New("packet damaged or lost in transmit")
	}

	return err

}

func (lu *LoraUtils) SetPreamble(length uint16) []byte {
	msbPreamble := byte(length>>8) & 0xff
	lsbPreamble := byte(length>>0) & 0xff
	return []byte{msbPreamble, lsbPreamble}
}

func (lu *LoraUtils) SetCrc(enable bool, currentModemConfig byte) byte {
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

func (lu *LoraUtils) SetWriteMask(reg byte) byte {
	return reg | 0x80

}

func (lu *LoraUtils) SetReadMask(reg byte) byte {
	return reg & 0x7f
}

func (lu *LoraUtils) ChangeMode(mode LoraMode) byte {
	loraMode := MODE_LONG_RANGE_MODE
	var selectedMode byte = 0
	switch mode {
	case Sleep:
		selectedMode = loraMode | MODE_SLEEP
	case Idle:
		selectedMode = loraMode | MODE_STDBY
	case Tx:
		selectedMode = loraMode | MODE_TX
	case RxContinuous:
		selectedMode = loraMode | MODE_RX_CONTINUOUS
	case RxSingle:
		selectedMode = loraMode | MODE_RX_SINGLE
	default:
		selectedMode = loraMode | MODE_STDBY
	}
	return selectedMode
}

func (lu *LoraUtils) SetTxPower(power byte) byte {
	return power | PA_BOOST
}

func (lu *LoraUtils) SetFreq(freq uint64) []byte {
	msb := byte(freq>>16) & 0xff
	mid := byte(freq>>8) & 0xff
	lsb := byte(freq) & 0xff
	return []byte{msb, mid, lsb}
}

func (lu *LoraUtils) SetSF(sf byte) byte {
	sfReg := sf << 4 & 0xf0
	return sfReg
}

func (lu *LoraUtils) SetBW(bw byte) byte {
	bwReg := bw << 4 & 0xf0
	return bwReg
}
