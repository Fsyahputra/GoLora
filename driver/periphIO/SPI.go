package periphIO

import (
	"errors"
	"log"
	"regexp"
	"sync"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
)

type SpiConf struct {
	Freq   physic.Frequency
	Mode   spi.Mode
	Bit    int
	CSName string
	CSSoft bool
	Reg    string
}

type SPI struct {
	Reg       string
	SpiDev    spi.Conn
	SpiCloser spi.PortCloser
	*SpiConf
	CSPin gpio.PinIO
	mu    sync.Mutex
}

func NewDefaultConf() *SpiConf {
	return &SpiConf{
		Freq:   100 * physic.KiloHertz,
		Mode:   spi.Mode0,
		Bit:    8,
		Reg:    "/dev/spidev0.0",
		CSName: "",
		CSSoft: false,
	}
}

func NewSpiConf(conf *SpiConf) (*SpiConf, error) {
	defConf := NewDefaultConf()
	if conf.Freq != 0 {
		defConf.Freq = conf.Freq
	}
	if conf.Mode != 0 {
		defConf.Mode = conf.Mode
	}
	if conf.Bit != 0 {
		defConf.Bit = conf.Bit
	}
	if conf.CSName == "" && conf.CSSoft == true {
		defConf.CSName = conf.CSName
		defConf.CSSoft = conf.CSSoft
		return nil, errors.New("CSName or CSSoft is required")
	}
	if conf.Reg != "" {
		defConf.Reg = conf.Reg
	}
	re := regexp.MustCompile(`^/dev/spidev[0-9]+\.[0-9]+$`)

	if !re.MatchString(defConf.Reg) {
		return nil, errors.New("invalid Register")
	}
	return defConf, nil
}

func NewSPI(spiConf *SpiConf) (*SPI, error) {

	if spiConf == nil {
		return nil, errors.New("spi conf is nil")
	}
	return &SPI{
		SpiDev:    nil,
		SpiCloser: nil,
		SpiConf:   spiConf,
		CSPin:     nil,
		mu:        sync.Mutex{},
	}, nil
}

func (pi *SPI) checkSpiDev() error {
	if pi.SpiDev == nil {
		return errors.New("no spi device connected")
	}
	return nil
}

func (pi *SPI) checkSpiCloser() error {
	if pi.SpiCloser == nil {
		return errors.New("no spi closer connected")
	}
	return nil
}

func (pi *SPI) Init() error {
	p, err := spireg.Open(pi.Reg)
	if err != nil {
		return err
	}
	conn, err := p.Connect(pi.Freq, pi.Mode, pi.Bit)
	if err != nil {
		return err
	}
	pi.SpiCloser = p
	pi.SpiDev = conn
	return nil
}

func (pi *SPI) closeConn() {
	if pi.SpiCloser != nil {
		pi.SpiCloser.Close()
	}
}

func (pi *SPI) checkSpiDevAndCloser() {
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

func (pi *SPI) softTx(reg, value byte) (byte, error) {
	pi.CSPin.Out(gpio.Low)
	defer pi.CSPin.Out(gpio.High)
	tx := []byte{reg, value}
	rx := make([]byte, len(tx))
	if err := pi.SpiDev.Tx(tx, rx); err != nil {
		return 0, err
	}
	return rx[1], nil
}

func (pi *SPI) tx(reg, value byte) (byte, error) {
	tx := []byte{reg, value}
	rx := make([]byte, len(tx))
	if err := pi.SpiDev.Tx(tx, rx); err != nil {
		return 0, err
	}
	return rx[1], nil
}

func (pi *SPI) softCSTx(reg, value byte) error {
	_, err := pi.softTx(reg, value)
	if err != nil {
		return err
	}
	return nil
}

func (pi *SPI) softCSRx(reg byte) (byte, error) {
	rx, err := pi.softTx(reg, 0)
	if err != nil {
		return 0, err
	}
	return rx, nil
}

func (pi *SPI) hardCSTx(reg, value byte) error {
	if _, err := pi.tx(reg, value); err != nil {
		return err
	}
	return nil
}

func (pi *SPI) hardCSRx(reg byte) (byte, error) {
	rx, err := pi.tx(reg, 0)
	if err != nil {
		return 0, err
	}
	return rx, nil
}

func (pi *SPI) SendToMod(reg, value byte) error {
	pi.checkSpiDevAndCloser()
	pi.mu.Lock()
	defer pi.mu.Unlock()
	var err error
	if pi.CSSoft {
		err = pi.softCSTx(reg, value)
	} else {
		err = pi.hardCSTx(reg, value)
	}
	return err
}

func (pi *SPI) ReadFromMod(reg byte) (byte, error) {
	pi.checkSpiDevAndCloser()
	pi.mu.Lock()
	defer pi.mu.Unlock()
	var err error
	var rx byte
	if pi.CSSoft {
		rx, err = pi.softCSRx(reg)
	} else {
		rx, err = pi.hardCSRx(reg)
	}
	return rx, err
}
