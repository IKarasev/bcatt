package emulator

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/IKarasev/bcatt/internal/globals"
)

var (
	HTTP_ADDR            string = "127.0.0.1"
	HTTP_PORT            string = "8080"
	RSS_READ_UPDATE_TIME        = time.Millisecond * 100
	OP_PAUSE_MILISEC            = time.Millisecond * 500
	WITH_LOG                    = false
	START_UTXO           globals.StartUtxo
)

func InitSettings(cf *globals.BcattConfig) error {
	if cf != nil {
		SetEmulatorSettings(cf)
		return nil
	}
	return LoadSettingsFromEnv()
}

// Loads web server settings from Enviroment variables
//
// BCATT_HTTP_ADDR  - ip address for web server to listen to
//
// BCATT_HTTP_PORT  - port for web server to listen
//
// BCATT_RSS_UPDATE - send update period in Milliseconds for RSS messages
func LoadSettingsFromEnv() error {
	errStr := ""
	if v := os.Getenv("BCATT_HTTP_ADDR"); v != "" {
		HTTP_ADDR = v
	}
	if v := os.Getenv("BCATT_HTTP_PORT"); v != "" {
		if _, err := strconv.Atoi(v); err != nil {
			errStr += "Failed to paser HTTP_PORT env variable\n"
		}
		HTTP_PORT = v
	}
	if v := os.Getenv("BCATT_RSS_UPDATE"); v != "" {
		if c, err := strconv.Atoi(v); err == nil {
			RSS_READ_UPDATE_TIME = time.Millisecond * time.Duration(c)
		} else {
			errStr += "Failed to pase BCATT_RSS_UPDATE env variable\n"
		}
	}
	if v := os.Getenv("OP_PAUSE_MILISEC"); v != "" {
		if c, err := strconv.Atoi(v); err == nil {
			OP_PAUSE_MILISEC = time.Millisecond * time.Duration(c)
		} else {
			errStr += "Failed to pase OP_PAUSE_MILISEC env variable\n"
		}
	}
	if v := os.Getenv("WITH_LOG"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			WITH_LOG = b
		} else {
			errStr += "Failed to pase WITH_LOG env variable\n"
		}
	}
	if errStr != "" {
		return fmt.Errorf("Load Emulator settings from ENV: failed to parse some settings, using default values:\n%s", errStr)
	}
	return nil

}

func SetEmulatorSettings(cf *globals.BcattConfig) {
	HTTP_ADDR = cf.Web.Address
	HTTP_PORT = cf.Web.Port
	if cf.Web.RssUpdateTime > 200 {
		RSS_READ_UPDATE_TIME = time.Millisecond * time.Duration(cf.Web.RssUpdateTime)
	}
	if cf.Emulator.OpPause >= 0 {
		OP_PAUSE_MILISEC = time.Microsecond * time.Duration(cf.Emulator.OpPause)
	}
	WITH_LOG = cf.Emulator.WithLog
	START_UTXO = cf.Emulator.StartUtxo
}
