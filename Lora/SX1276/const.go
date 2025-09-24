package SX1276

import (
	"github.com/Fsyahputra/GoLora/Lora/SX1276/internal"
)

type BW uint64

const (
	BW_1 BW = 10.4e3
	BW_2 BW = 15.6e3
	BW_3 BW = 20.8e3
	BW_4 BW = 31.24e3
	BW_5 BW = 41.7e3
	BW_6 BW = 62.5e3
	BW_7 BW = 125e3
	BW_8 BW = 250e3
)

type Header bool

const (
	Explicit Header = true
	Implicit Header = false
)

type Event int

const (
	OnRxDone = iota
	OnTxDone
)

const (
	Sleep        = LoraMode(internal.MODE_SLEEP)
	Idle         = LoraMode(internal.MODE_STDBY)
	Tx           = LoraMode(internal.MODE_TX)
	RxSingle     = LoraMode(internal.MODE_RX_SINGLE)
	RxContinuous = LoraMode(internal.MODE_RX_CONTINUOUS)
)

type LoraMode byte
