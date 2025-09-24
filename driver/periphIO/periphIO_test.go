package periphIO

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

const (
	RSTPINMOD0 = "GPIO36"
	RSTPINMOD1 = "GPIO38"
	CBPINMOD0  = "GPIO133"
	CBPINMOD1  = "GPIO134"
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

// MODULE0 is CONNECTED TO GPIO36 (RST) AND SPIDEV0.0
// MODULE1 is CONNECTED TO GPIO38 (RST) AND SPIDEV4.0
func TestRSTPin_HighLow(t *testing.T) {
	initHost()
	mod0rstPin, err := NewRstPinPeriphIO("GPIO36")
	if err != nil {
		t.Fatalf("Failed to create RSTPin: %v", err)
	}
	mod0Reader := gpioreg.ByName("GPIO133")
	if mod0Reader == nil {
		t.Fatalf("Failed to find GPIO133")
	}
	mod0rstPin.pin.Out(gpio.Low) // Ensure starting state is Low
	mod0Reader.In(gpio.PullNoChange, gpio.BothEdges)
	tests := []struct {
		name   string
		action func() error
		want   bool
	}{
		{
			name:   "Set High module 0",
			action: mod0rstPin.High,
			want:   true,
		},
		{
			name:   "Set Low module 0",
			action: mod0rstPin.Low,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action()
			if err != nil {
				t.Fatalf("Action failed: %v", err)
			}
			time.Sleep(1000 * time.Millisecond) // Small delay to ensure the state change is registered
			val := mod0Reader.Read()
			if tt.want {
				assert.Equal(t, gpio.High, val, "Expected pin to be High")
			} else {
				assert.Equal(t, gpio.Low, val, "Expected pin to be Low")
			}
		})
	}

	mod1rstPin, err := NewRstPinPeriphIO("GPIO38")
	if err != nil {
		t.Fatalf("Failed to create RSTPin: %v", err)
	}
	mod1Reader := gpioreg.ByName("GPIO134")
	if mod1Reader == nil {
		t.Fatalf("Failed to find GPIO134")
	}
	mod1rstPin.pin.Out(gpio.Low) // Ensure starting state is Low
	mod1Reader.In(gpio.PullNoChange, gpio.BothEdges)
	tests2 := []struct {
		name   string
		action func() error
		want   bool
	}{
		{
			name:   "Set High module 1",
			action: mod1rstPin.High,
			want:   true,
		},
		{
			name:   "Set Low module 1",
			action: mod1rstPin.Low,
			want:   false,
		},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action()
			if err != nil {
				t.Fatalf("Action failed: %v", err)
			}
			time.Sleep(1000 * time.Millisecond) // Small delay to ensure the state change is registered
			val := mod1Reader.Read()
			if tt.want {
				assert.Equal(t, gpio.High, val, "Expected pin to be High")
			} else {
				assert.Equal(t, gpio.Low, val, "Expected pin to be Low")
			}
		})
	}

}

func readValTest(cbPinName, testPinName string, level gpio.Level, want bool) func(t *testing.T) {
	return func(t *testing.T) {
		cbPin, err := NewCbPin(cbPinName)
		if err != nil {
			t.Fatalf("Failed to create CbPin: %v", err)
		}

		err = cbPin.Init()
		if err != nil {
			t.Fatalf("Failed to initialize CbPin: %v", err)
		}
		testPin := gpioreg.ByName(testPinName)
		if testPin == nil {
			t.Fatalf("Failed to find %s", testPinName)
		}
		err = testPin.Out(level)
		if err != nil {
			t.Fatalf("Failed to set %s: %v", testPinName, err)
		}
		time.Sleep(10 * time.Millisecond)
		val, err := cbPin.ReadVal()
		if err != nil {
			t.Fatalf("ReadVal failed: %v", err)
		}
		assert.Equal(t, want, val, "ReadVal did not return expected value")
	}
}
func ReadVal(mod int) (testFunc []func(t *testing.T), names []string, err error) {
	initHost()
	var cbPinName, testPinName string
	if mod == 0 {
		cbPinName = CBPINMOD0
		testPinName = RSTPINMOD0
	} else if mod == 1 {
		cbPinName = CBPINMOD1
		testPinName = RSTPINMOD1
	} else {
		return nil, nil, errors.New("Invalid module number")
	}

	return []func(t *testing.T){
		readValTest(cbPinName, testPinName, gpio.High, true),
		readValTest(cbPinName, testPinName, gpio.Low, false),
	}, []string{fmt.Sprintf("MOD %v it should read High", mod), fmt.Sprintf("MOD %v it should read Low", mod)}, nil
}

func TestCbPin_ReadVal(t *testing.T) {
	for mod := 0; mod <= 1; mod++ {
		testFuncs, names, err := ReadVal(mod)
		if err != nil {
			t.Fatalf("Setup for MOD %v failed: %v", mod, err)
		}
		for i, testFunc := range testFuncs {
			t.Run(names[i], testFunc)
		}
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
			_, err := NewDriver(CBPINMOD0, RSTPINMOD0, tt.spiConf)
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

func ReadMod(mod int) (func(t *testing.T), error) {
	initHost()
	conf, rstPinName, err := getConfForMod(mod)
	if err != nil {
		return nil, err
	}
	rstPin := gpioreg.ByName(rstPinName)
	if rstPin == nil {
		return nil, errors.New("Failed to find " + rstPinName)
	}
	resetMod(rstPin)
	spi, err := NewSPI(conf)
	if err != nil {
		return nil, errors.New("Failed to create SPI: " + err.Error())
	}
	err = spi.Init()
	if err != nil {
		return nil, errors.New("Failed to initialize SPI: " + err.Error())
	}
	tests := struct {
		name   string
		input  byte
		expect byte
	}{
		input:  0x42,
		expect: 0x12,
	}
	return func(t *testing.T) {
		val, err := spi.ReadFromMod(tests.input)
		if err != nil {
			t.Fatalf("ReadFromMod failed: %v", err)
		}
		assert.Equal(t, tests.expect, val, "ReadFromMod did not return expected value")
		//spi.CloseConn()
	}, nil
}

func getConfForMod(mod int) (*SpiConf, string, error) {
	defSpiConf := NewDefaultConf()
	if mod == 0 {
		defSpiConf.Reg = "/dev/spidev0.0"
		return defSpiConf, RSTPINMOD0, nil
	} else if mod == 1 {
		defSpiConf.Reg = "/dev/spidev4.0"
		return defSpiConf, RSTPINMOD1, nil
	}
	return nil, "", errors.New("Invalid module number")
}

func TestSPI_ReadFromMod(t *testing.T) {
	tests := []struct {
		name string
		mod  int
	}{
		{
			name: "it Should read from module 0",
			mod:  0,
		},
		{
			name: "it Should read from module 1",
			mod:  1,
		},
	}

	for _, tt := range tests {
		testFunc, err := ReadMod(tt.mod)
		if err != nil {
			t.Fatalf("Setup for %s failed: %v", tt.name, err)
		}
		t.Run(tt.name, testFunc)
	}
}

func SendToMod(mod int) ([]func(t *testing.T), error) {
	initHost()

	conf, rstPinName, err := getConfForMod(mod)
	if err != nil {
		return nil, err
	}

	rstPin := gpioreg.ByName(rstPinName)
	if rstPin == nil {
		return nil, fmt.Errorf("failed to find %s", rstPinName)
	}

	resetMod(rstPin)

	spi, err := NewSPI(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create SPI: %w", err)
	}

	if err := spi.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize SPI: %w", err)
	}

	testFunc1 := func(t *testing.T) {
		expected := byte(0x09)
		val, err := spi.ReadFromMod(0x01)
		if err != nil {
			t.Fatalf("ReadFromMod failed: %v", err)
		}
		assert.Equal(t, expected, val, "ReadFromMod did not return expected value")
	}

	testFunc2 := func(t *testing.T) {
		writeVal := byte(0x00)
		reg := byte(0x01) | 0x80 // masked for write
		if err := spi.SendToMod(reg, writeVal); err != nil {
			t.Fatalf("SendToMod failed: %v", err)
		}
		val, err := spi.ReadFromMod(0x01)
		if err != nil {
			t.Fatalf("ReadFromMod failed: %v", err)
		}
		assert.Equal(t, writeVal, val, "ReadFromMod did not return expected value after SendToMod")
	}

	testFunc3 := func(t *testing.T) {
		resetMod(rstPin)
		expected := byte(0x09)
		val, err := spi.ReadFromMod(0x01)
		if err != nil {
			t.Fatalf("ReadFromMod failed: %v", err)
		}
		assert.Equal(t, expected, val, "ReadFromMod did not return expected value after reset")
	}

	return []func(t *testing.T){testFunc1, testFunc2, testFunc3}, nil
}

func TestSPI_SendToMod(t *testing.T) {
	tests := []struct {
		name string
		mod  int
	}{
		{
			name: "it Should send to module 0",
			mod:  0,
		},
		{
			name: "it Should send to module 1",
			mod:  1,
		},
	}

	for _, tt := range tests {
		testFuncs, err := SendToMod(tt.mod)
		if err != nil {
			t.Fatalf("Setup for %s failed: %v", tt.name, err)
		}
		for _, testFunc := range testFuncs {
			t.Run(tt.name, testFunc)
		}
	}
}
