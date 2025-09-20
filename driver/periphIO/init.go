package periphIO

import (
	"github.com/Fsyahputra/GoLora/driver"
)

type PeriphDriver struct {
	CbPin  *CbPin
	RSTPin *RSTPin
	SPI    *SPI
}

func NewDriver(CbPinName, RstPinName string, conf *SpiConf) (*PeriphDriver, error) {
	HwCbPin, err := NewCbPin(CbPinName)
	if err != nil {
		return nil, err
	}
	HwRstPin, err := NewRstPinPeriphIO(RstPinName)
	if err != nil {
		return nil, err
	}
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	HwSpi, err := NewSPI(conf)
	if err != nil {
		return nil, err
	}
	return &PeriphDriver{
		CbPin:  HwCbPin,
		RSTPin: HwRstPin,
		SPI:    HwSpi,
	}, nil
}

func (d *PeriphDriver) Init() (*driver.Driver, error) {
	if err := d.CbPin.Init(); err != nil {
		return nil, err
	}
	if err := d.SPI.Init(); err != nil {
		return nil, err
	}
	newDrv := &driver.Driver{
		RSTPin:  d.RSTPin,
		CbPin:   d.CbPin,
		ModComm: d.SPI,
	}

	return newDrv, nil
}
