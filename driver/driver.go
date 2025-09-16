package driver

type ModComm interface {
	SendToMod(reg, value byte) error
	ReadFromMod(reg byte) (byte, error)
}

type RSTPin interface {
	Low() error
	High() error
}

type CbPin interface {
	ReadVal() (bool, error)
}

type HwDriver interface {
	Init() error
}
