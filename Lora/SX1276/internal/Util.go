package internal

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
}

type LoraUtils struct{}

func (lu *LoraUtils) SetWriteMask(reg byte) byte {

}

func (lu *LoraUtils) SetReadMask(reg byte) byte {

}

func (lu *LoraUtils) ChangeMode(mode LoraMode) byte {

}
