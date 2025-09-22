package periphIO

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
	Bit    uint
	CSName string
	CSSoft bool
	Reg    string
}

type SPI struct {
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

func (c *SpiConf) Validate() error {
	if c == nil {
		return errors.New("spi conf is nil")
	}
	if c.Freq <= physic.Frequency(0) {
		return errors.New("freq must be greater than 0")
	}
	if c.Bit != 8 && c.Bit != 16 {
		return errors.New("bit must be 8, or 16")
	}

	if c.Mode < 0 || c.Mode > 3 {
		return errors.New("mode must be 0, 1, 2, or 3")
	}

	if c.CSName == "" && c.CSSoft == true {
		return errors.New("you Should provide CSName when using software CS control")
	}

	if c.CSName != "" && c.CSSoft == false {
		return errors.New("you Shouldn't provide CSName when using hardware CS control")
	}
	re := regexp.MustCompile(`^/dev/spidev[0-9]+\.[0-9]+$`)

	if !re.MatchString(c.Reg) {
		return errors.New("invalid Register")
	}
	return nil
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
	p, err := spireg.Open(pi.SpiConf.Reg)
	if err != nil {
		return errors.New("spireg: can't open unknown port")
	}
	conn, err := p.Connect(pi.Freq, pi.Mode, int(pi.Bit))
	if err != nil {
		return err
	}
	pi.SpiCloser = p
	pi.SpiDev = conn

	if !pi.CSSoft {
		return nil
	}
	csp := gpioreg.ByName(pi.CSName)
	if csp == nil {
		return errors.New("CsPin GPIO does not exist")
	}
	pi.CSPin = csp
	return nil
}

func (pi *SPI) CloseConn() {
	if pi.SpiCloser != nil {
		_ = pi.SpiCloser.Close()
	}
}

func (pi *SPI) checkSpiDevAndCloser() {
	err := pi.checkSpiDev()
	if err != nil {
		pi.CloseConn()
		log.Fatal(err.Error())
	}
	err = pi.checkSpiCloser()
	if err != nil {
		pi.CloseConn()
		log.Fatal(err.Error())
	}
}

func (pi *SPI) softTx(reg, value byte) (byte, error) {
	err := pi.CSPin.Out(gpio.Low)
	if err != nil {
		return 0, err
	}
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
