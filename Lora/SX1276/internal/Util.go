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
	SetFreq(freq byte) []byte
	SetSF(sf byte) byte
	SetBW(bw byte) byte
	SetCrc(enable bool) byte
	SetPreamble(length uint16) []byte
	SetSyncWord(reg byte) byte
	CheckData(irq byte) error
	GetPktLenByHeader(header bool) (uint8, error)
	SetCodingRate(cr byte, currentModemConfig byte) byte
}

type LoraUtils struct{}

func (lu *LoraUtils) SetCodingRate(cr byte, currentModemConfig byte) byte {

}

func (lu *LoraUtils) GetPktLenByHeader(header bool) (uint8, error) {

}

func (lu *LoraUtils) CheckData(irq byte) error {
	var err error = nil
	if irq&IRQ_RX_DONE_MASK == 0 {
		err = errors.New("no Packet Received")
	}

	if irq&IRQ_PAYLOAD_CRC_ERROR_MASK == 0 {
		err = errors.New("packet Rusak")
	}
	return err

}

func (lu *LoraUtils) SetSyncWord(reg byte) byte {

}

func (lu *LoraUtils) SetPreamble(length uint16) []byte {

}

func (lu *LoraUtils) SetCrc(enable bool) byte {
	//TODO implement me
	panic("implement me")
}

func (lu *LoraUtils) SetWriteMask(reg byte) byte {

}

func (lu *LoraUtils) SetReadMask(reg byte) byte {

}

func (lu *LoraUtils) ChangeMode(mode LoraMode) byte {}

func (lu *LoraUtils) SetTxPower(power byte) byte {}

func (lu *LoraUtils) SetFreq(freq byte) []byte {

}

func (lu *LoraUtils) SetSF(sf byte) byte {

}

func (lu *LoraUtils) SetBW(bw byte) byte {

}
