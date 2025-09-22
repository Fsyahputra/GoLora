package periphIO

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

func initHost() {
	_, err := host.Init()
	if err != nil {
		panic("Failed to initialize periph.io: " + err.Error())
	}
}

func TestNewRstPinPeriphIO(t *testing.T) {
	initHost()
	tests := []struct {
		name    string
		pinName string
		want    error
	}{
		{
			name:    "Valid pin name",
			pinName: "GPIO133",
			want:    nil,
		},
		{
			name:    "Invalid pin name",
			pinName: "INVALID_PIN",
			want:    errors.New("GPIO does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRstPinPeriphIO(tt.pinName)
			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.want.Error())
			}
		})
	}
}

func TestRSTPin_HighLow(t *testing.T) {
	initHost()
	rstPin, err := NewRstPinPeriphIO("GPIO134")
	if err != nil {
		t.Fatalf("Failed to create RSTPin: %v", err)
	}
	reader := gpioreg.ByName("GPIO133")
	if reader == nil {
		t.Fatalf("Failed to find GPIO133")
	}
	rstPin.pin.Out(gpio.Low) // Ensure starting state is Low
	reader.In(gpio.PullNoChange, gpio.BothEdges)
	tests := []struct {
		name   string
		action func() error
		want   bool
	}{
		{
			name:   "Set High",
			action: rstPin.High,
			want:   true,
		},
		{
			name:   "Set Low",
			action: rstPin.Low,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action()
			if err != nil {
				t.Fatalf("Action failed: %v", err)
			}
			val := reader.Read()
			if tt.want {
				assert.Equal(t, gpio.High, val, "Expected pin to be High")
			} else {
				assert.Equal(t, gpio.Low, val, "Expected pin to be Low")
			}
		})
	}
}

func TestCbPin_ReadVal(t *testing.T) {
	initHost()
	cbPin, err := NewCbPin("GPIO134")
	if err != nil {
		t.Fatalf("Failed to create CbPin: %v", err)
	}

	err = cbPin.Init()
	if err != nil {
		t.Fatalf("Failed to initialize CbPin: %v", err)
	}
	p := gpioreg.ByName("GPIO133")
	if p == nil {
		t.Fatalf("Failed to find GPIO133")
	}
	err = p.Out(gpio.Low)
	if err != nil {
		t.Fatalf("Failed to set GPIO133 to Low: %v", err)
	}

	tests := []struct {
		name string
		want gpio.Level
	}{
		{
			name: "ReadVal Must Be High",
			want: gpio.High,
		},
		{
			name: "ReadVal Must Be Low",
			want: gpio.Low,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = p.Out(tt.want)
			if err != nil {
				t.Fatalf("Failed to set GPIO133: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
			val, err := cbPin.ReadVal()
			if err != nil {
				t.Fatalf("ReadVal failed: %v", err)
			}
			assert.Equal(t, bool(tt.want), val, "ReadVal did not return expected value")
		})
	}
}

func TestNewCbPin(t *testing.T) {
	initHost()
	tests := []struct {
		name    string
		namePin string
		want    error
	}{
		{
			name:    "Valid pin name",
			namePin: "GPIO133",
			want:    nil,
		},
		{
			name:    "Invalid pin name",
			namePin: "INVALID_PIN",
			want:    errors.New("GPIO does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCbPin(tt.namePin)
			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.want.Error())
			}
		})
	}
}

func TestNewDriver(t *testing.T) {
	initHost()
	tests := []struct {
		name    string
		rstPin  string
		cbPin   string
		wantErr error
	}{
		{
			name:    "it Should return driver if both pins are valid",
			rstPin:  "GPIO134",
			cbPin:   "GPIO133",
			wantErr: nil,
		},
		{
			name:    "it Should return error if rstPin is invalid",
			rstPin:  "INVALID_PIN",
			cbPin:   "GPIO133",
			wantErr: errors.New("GPIO does not exist"),
		},
		{
			name:    "it Should return error if cbPin is invalid",
			rstPin:  "GPIO134",
			cbPin:   "INVALID_PIN",
			wantErr: errors.New("GPIO does not exist"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDriver(tt.rstPin, tt.cbPin, NewDefaultConf())
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			}
		})
	}

	tests2 := []struct {
		name    string
		spiConf *SpiConf
		want    error
	}{
		{
			name: "it Should return driver if spiConf is valid",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   0,
				Bit:    8,
				CSName: "",
				CSSoft: false,
				Reg:    "/dev/spidev0.0",
			},
			want: nil,
		},
		{
			name:    "it Should return driver with default conf if spiConf is nil",
			spiConf: nil,
			want:    errors.New("spi conf is nil"),
		},
		{
			name: "it Should return error if spiConf is invalid (frequency negative)",
			spiConf: &SpiConf{
				Freq:   -100 * physic.KiloHertz,
				Mode:   0,
				Bit:    8,
				CSName: "",
				CSSoft: false,
				Reg:    "/dev/spidev0.0",
			},
			want: errors.New("freq must be greater than 0"),
		},
		{
			name: "it Should return error if spiConf is invalid (bit not 8 or 16)",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   0,
				Bit:    7,
				CSName: "",
				CSSoft: false,
				Reg:    "/dev/spidev0.0",
			},
			want: errors.New("bit must be 8, or 16"),
		},
		{
			name: "it Should return error if spiConf is invalid (mode not 0-3)",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   4,
				Bit:    8,
				CSName: "",
				CSSoft: false,
				Reg:    "/dev/spidev0.0",
			},
			want: errors.New("mode must be 0, 1, 2, or 3"),
		},
		{
			name: "it Should return error if spiConf is invalid (mode < 0)",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   -1,
				Bit:    8,
				CSName: "",
				CSSoft: false,
				Reg:    "/dev/spidev0.0",
			},
			want: errors.New("mode must be 0, 1, 2, or 3"),
		},
		{
			name: "it Should return error if spiConf is invalid (reg not matching pattern)",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   0,
				Bit:    8,
				CSName: "",
				CSSoft: false,
				Reg:    "invalid_reg",
			},
			want: errors.New("invalid Register"),
		},
		{
			name: "it Should return error if spiConf is invalid (CSName is empty and CSSoft is true)",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   0,
				Bit:    8,
				CSName: "",
				CSSoft: true,
				Reg:    "/dev/spidev0.0",
			},
			want: errors.New("you Should provide CSName when using software CS control"),
		},
		{
			name: "it Should return error if spiConf is invalid (CSName does not exist)",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   0,
				Bit:    8,
				CSName: "",
				CSSoft: true,
				Reg:    "/dev/spidev0.0",
			},
			want: errors.New("you Should provide CSName when using software CS control"),
		},
		{
			name: "it Should return error if spiConf is invalid (CSName is provided and CSSoft is false)",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   0,
				Bit:    8,
				CSName: "GPIO133",
				CSSoft: false,
				Reg:    "/dev/spidev0.0",
			},
			want: errors.New("you Shouldn't provide CSName when using hardware CS control"),
		},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDriver("GPIO34", "GPIO35", tt.spiConf)
			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.want.Error())
			}
		})
	}
}

func TestSPI_Init(t *testing.T) {
	initHost()
	defSpiConf := NewDefaultConf()
	tests := []struct {
		name    string
		spiConf func() SpiConf
		wantErr error
	}{
		{
			name:    "it Should return nil if spiConf is valid",
			spiConf: func() SpiConf { return *defSpiConf },
			wantErr: nil,
		},
		{
			name: "it Should return error if reg is invalid",
			spiConf: func() SpiConf {
				c := *defSpiConf
				c.Reg = "invalid_reg"
				return c
			},
			wantErr: errors.New("spireg: can't open unknown port"),
		},
		{
			name: "it Should return error if csName does not exist",
			spiConf: func() SpiConf {
				c := *defSpiConf
				c.CSName = "INVALID_PIN"
				c.CSSoft = true
				return c
			},
			wantErr: errors.New("CsPin GPIO does not exist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newConf := tt.spiConf()
			spi, err := NewSPI(&newConf)
			defer spi.CloseConn()
			assert.NoError(t, err)
			err = spi.Init()
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			}
		})
	}

}

func resetMod(p gpio.PinIO) {
	p.Out(gpio.Low)
	time.Sleep(500 * time.Millisecond)
	p.Out(gpio.High)
	time.Sleep(500 * time.Millisecond)
}

func TestSPI_ReadFromMod(t *testing.T) {
	initHost()
	defSpiConf := NewDefaultConf()
	p := gpioreg.ByName("GPIO46")
	if p == nil {
		t.Fatalf("Failed to find GPIO46")
	}
	resetMod(p)
	spi, err := NewSPI(defSpiConf)
	if err != nil {
		t.Fatalf("Failed to create SPI: %v", err)
	}
	defer spi.CloseConn()
	err = spi.Init()
	if err != nil {
		t.Fatalf("Failed to initialize SPI: %v", err)
	}
	tests := struct {
		name   string
		input  byte
		expect byte
	}{
		name:   "it Should read the version from the module",
		input:  0x42,
		expect: 0x12,
	}
	t.Run(tests.name, func(t *testing.T) {
		val, err := spi.ReadFromMod(tests.input)
		if err != nil {
			t.Fatalf("ReadFromMod failed: %v", err)
		}
		assert.Equal(t, tests.expect, val, "ReadFromMod did not return expected value")
	})
}

func TestSPI_SendToMod(t *testing.T) {
	initHost()
	defSpiConf := NewDefaultConf()
	p := gpioreg.ByName("GPIO46")
	if p == nil {
		t.Fatalf("Failed to find GPIO46")
	}
	resetMod(p)

	spi, err := NewSPI(defSpiConf)
	if err != nil {
		t.Fatalf("Failed to create SPI: %v", err)
	}
	err = spi.Init()
	if err != nil {
		t.Fatalf("Failed to initialize SPI: %v", err)
	}
	tests := struct {
		name        string
		reg         byte
		expectedVal byte
	}{
		name:        "it Should read modemConf1 default value before write anything to it",
		reg:         0x01,
		expectedVal: 0x09, // Default value of modemConf1 register
	}
	t.Run(tests.name, func(t *testing.T) {
		val, err := spi.ReadFromMod(tests.reg)
		if err != nil {
			t.Fatalf("ReadFromMod failed: %v", err)
		}
		assert.Equal(t, tests.expectedVal, val, "ReadFromMod did not return expected value")
	})

	test2 := struct {
		name        string
		reg         byte
		writeVal    byte
		expectedVal byte
	}{
		name:        "it Should write to modemConf1 and read back the same value",
		reg:         0x01,
		writeVal:    0x00,
		expectedVal: 0x00,
	}

	t.Run(test2.name, func(t *testing.T) {
		maskedReg := test2.reg | 0x80
		err := spi.SendToMod(maskedReg, test2.writeVal)
		if err != nil {
			t.Fatalf("SendToMod failed: %v", err)
		}
		val, err := spi.ReadFromMod(test2.reg)
		if err != nil {
			t.Fatalf("ReadFromMod failed: %v", err)
		}
		assert.Equal(t, test2.expectedVal, val, "ReadFromMod did not return expected value after SendToMod")
	})

	resetMod(p)
	tests3 := struct {
		name        string
		reg         byte
		expectedVal byte
	}{
		name:        "it Should read modemConf1 default value before write anything to it after reset",
		reg:         0x01,
		expectedVal: 0x09, // Default value of modemConf1 register
	}
	t.Run(tests3.name, func(t *testing.T) {
		val, err := spi.ReadFromMod(tests3.reg)
		if err != nil {
			t.Fatalf("ReadFromMod failed: %v", err)
		}
		assert.Equal(t, tests3.expectedVal, val, "ReadFromMod did not return expected value")
	})
}
