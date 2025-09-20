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
			name: "it Should use default reg if spiConf is invalid (Reg empty)",
			spiConf: &SpiConf{
				Freq:   100 * physic.KiloHertz,
				Mode:   0,
				Bit:    8,
				CSName: "",
				CSSoft: false,
				Reg:    "",
			},
			want: nil,
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
				CSName: "INVALID_PIN",
				CSSoft: true,
				Reg:    "/dev/spidev0.0",
			},
			want: errors.New("CsPin GPIO does not exist"),
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
