package GoLoraSpi

import (
	"errors"
	"log"
	"regexp"
	"sync"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
)

type SpiConf struct {
	Freq   physic.Frequency
	Mode   spi.Mode
	Bit    int
	CSPin  gpio.PinIO
	CSName string
	CSSoft bool
}

func (sc *SpiConf) Init() error {
	p := gpioreg.ByName(sc.CSName)
	if p == nil {
		return errors.New("CS not registered")
	}
	sc.CSPin = p
	return nil
}

type PeriphIOSPI struct {
	Reg       string
	SpiDev    spi.Conn
	SpiCloser spi.PortCloser
	SpiConf   *SpiConf
	mu        *sync.Mutex
}

func NewSpiConf(frequency physic.Frequency, mode spi.Mode, bit int) *SpiConf {
	defConf := &SpiConf{
		Freq: 100 * physic.KiloHertz,
		Mode: spi.Mode0,
		Bit:  8,
	}
	if frequency != 0 {
		defConf.Freq = frequency
	}
	if mode != 0 {
		defConf.Mode = mode
	}
	if bit != 0 {
		defConf.Bit = bit
	}
	return defConf
}

func NewPeriphIOSPI(reg string, spiConf *SpiConf) (*PeriphIOSPI, error) {
	re := regexp.MustCompile(`^/dev/spidev[0-9]+\.[0-9]+$`)

	if !re.MatchString(reg) {
		return nil, errors.New("invalid Register")
	}

	if spiConf == nil {
		return nil, errors.New("spi conf is nil")
	}
	return &PeriphIOSPI{
		Reg:       reg,
		SpiDev:    nil,
		SpiCloser: nil,
		SpiConf:   spiConf,
		mu:        &sync.Mutex{},
	}, nil
}

func (pi *PeriphIOSPI) checkSpiDev() error {
	if pi.SpiDev == nil {
		return errors.New("no spi device connected")
	}
	return nil
}

func (pi *PeriphIOSPI) checkSpiCloser() error {
	if pi.SpiCloser == nil {
		return errors.New("no spi closer connected")
	}
	return nil
}

func (pi *PeriphIOSPI) Init() error {
	if err := pi.SpiConf.Init(); err != nil {
		return err
	}

	p, err := spireg.Open(pi.Reg)
	if err != nil {
		return err
	}
	conn, err := p.Connect(pi.SpiConf.Freq, pi.SpiConf.Mode, pi.SpiConf.Bit)
	pi.SpiCloser = p
	pi.SpiDev = conn
	return nil
}

func (pi *PeriphIOSPI) closeConn() {
	if pi.SpiCloser != nil {
		pi.SpiCloser.Close()
	}
}

func (pi *PeriphIOSPI) checkSpiDevAndCloser() {
	err := pi.checkSpiDev()
	if err != nil {
		pi.closeConn()
		log.Fatal(err.Error())
	}
	err = pi.checkSpiCloser()
	if err != nil {
		pi.closeConn()
		log.Fatal(err.Error())
	}
}

func (pi *PeriphIOSPI) softTx(reg, value byte) (byte, error) {
	pi.SpiConf.CSPin.Out(gpio.Low)
	defer pi.SpiConf.CSPin.Out(gpio.High)
	tx := []byte{reg, value}
	rx := make([]byte, len(tx))
	if err := pi.SpiDev.Tx(tx, rx); err != nil {
		return 0, err
	}
	return rx[1], nil
}

func (pi *PeriphIOSPI) tx(reg, value byte) (byte, error) {
	tx := []byte{reg, value}
	rx := make([]byte, len(tx))
	if err := pi.SpiDev.Tx(tx, rx); err != nil {
		return 0, err
	}
	return rx[1], nil
}

func (pi *PeriphIOSPI) softCSTx(reg, value byte) error {
	_, err := pi.softTx(reg, value)
	if err != nil {
		return err
	}
	return nil
}

func (pi *PeriphIOSPI) softCSRx(reg byte) (byte, error) {
	rx, err := pi.softTx(reg, 0)
	if err != nil {
		return 0, err
	}
	return rx, nil
}

func (pi *PeriphIOSPI) hardCSTx(reg, value byte) error {
	if _, err := pi.tx(reg, value); err != nil {
		return err
	}
	return nil
}

func (pi *PeriphIOSPI) hardCSRx(reg byte) (byte, error) {
	rx, err := pi.tx(reg, 0)
	if err != nil {
		return 0, err
	}
	return rx, nil
}

func (pi *PeriphIOSPI) SendToMod(reg, value byte) error {
	pi.checkSpiDevAndCloser()
	pi.mu.Lock()
	defer pi.mu.Unlock()
	var err error
	if pi.SpiConf.CSSoft {
		err = pi.softCSTx(reg, value)
	} else {
		err = pi.hardCSTx(reg, value)
	}
	return err
}

func (pi *PeriphIOSPI) ReadFromMod(reg byte) (byte, error) {
	pi.checkSpiDevAndCloser()
	pi.mu.Lock()
	defer pi.mu.Unlock()
	var err error
	var rx byte
	if pi.SpiConf.CSSoft {
		rx, err = pi.softCSRx(reg)
	} else {
		rx, err = pi.hardCSRx(reg)
	}
	return rx, err
}
