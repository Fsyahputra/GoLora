package internal

// ============================
// Register definitions
// ============================
const (
	REG_FIFO                 byte = 0x00
	REG_OP_MODE              byte = 0x01
	REG_FRF_MSB              byte = 0x06
	REG_FRF_MID              byte = 0x07
	REG_FRF_LSB              byte = 0x08
	REG_PA_CONFIG            byte = 0x09
	REG_LNA                  byte = 0x0c
	REG_FIFO_ADDR_PTR        byte = 0x0d
	REG_FIFO_TX_BASE_ADDR    byte = 0x0e
	REG_FIFO_RX_BASE_ADDR    byte = 0x0f
	REG_FIFO_RX_CURRENT_ADDR byte = 0x10
	REG_IRQ_FLAGS_MASK       byte = 0x11
	REG_IRQ_FLAGS            byte = 0x12
	REG_RX_NB_BYTES          byte = 0x13
	REG_PKT_SNR_VALUE        byte = 0x19
	REG_PKT_RSSI_VALUE       byte = 0x1a
	REG_MODEM_CONFIG_1       byte = 0x1d
	REG_MODEM_CONFIG_2       byte = 0x1e
	REG_PREAMBLE_MSB         byte = 0x20
	REG_PREAMBLE_LSB         byte = 0x21
	REG_PAYLOAD_LENGTH       byte = 0x22
	REG_MODEM_CONFIG_3       byte = 0x26
	REG_RSSI_WIDEBAND        byte = 0x2c
	REG_DETECTION_OPTIMIZE   byte = 0x31
	REG_DETECTION_THRESHOLD  byte = 0x37
	REG_SYNC_WORD            byte = 0x39
	REG_DIO_MAPPING_1        byte = 0x40
	REG_VERSION              byte = 0x42
)

// ============================
// Transceiver modes
// ============================
const (
	MODE_LONG_RANGE_MODE byte = 0x80
	MODE_SLEEP           byte = 0x00
	MODE_STDBY           byte = 0x01
	MODE_TX              byte = 0x03
	MODE_RX_CONTINUOUS   byte = 0x05
	MODE_RX_SINGLE       byte = 0x06
)

// ============================
// PA configuration
// ============================
const PA_BOOST byte = 0x80

// ============================
// IRQ masks
// ============================
const (
	IRQ_TX_DONE_MASK           byte = 0x08
	IRQ_PAYLOAD_CRC_ERROR_MASK byte = 0x20
	IRQ_RX_DONE_MASK           byte = 0x40
)

// ============================
// PA output pins
// ============================
const (
	PA_OUTPUT_RFO_PIN      byte = 0
	PA_OUTPUT_PA_BOOST_PIN byte = 1
)
