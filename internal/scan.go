package internal

import (
	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

func Scan() {
	// Enable BLE interface.
	// Enable BLE interface.
	err := adapter.Enable()
	if err != nil {
		log.Fatalf("failed to enable ble stack: %v", err)
	}

	// Start scanning.
	log.Info("scanning...")
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		log.Infof("found device: %v %v %v %+v", device.Address.String(), device.RSSI, device.LocalName(), device)
		// println("found device:", device.Address.String(), device.RSSI, device.LocalName())
	})
	if err != nil {
		log.Fatalf("failed to scan for ble devices: %v", err)
	}
}
