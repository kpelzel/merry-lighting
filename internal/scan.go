package internal

import (
	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

func Scan() {
	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	// Start scanning.
	log.Info("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		log.Infof("found device: %v %v %v %+v", device.Address.String(), device.RSSI, device.LocalName(), device)
		// println("found device:", device.Address.String(), device.RSSI, device.LocalName())
	})
	must("start scan", err)
}
