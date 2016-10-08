package main

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hclog "github.com/brutella/hc/log"
	"github.com/sausheong/hs1xxplug"
)

var (
	plug hs1xxplug.Hs1xxPlug
	acc  *accessory.Switch
)

type systemInfoOutput struct {
	System struct {
		Sysinfo struct {
			ActiveMode string `json:"active_mode"`
			Alias      string `json:"alias"`
			DevName    string `json:"dev_name"`
			DeviceID   string `json:"deviceId"`
			ErrCode    int    `json:"err_code"`
			Feature    string `json:"feature"`
			FwID       string `json:"fwId"`
			HwID       string `json:"hwId"`
			HwVer      string `json:"hw_ver"`
			IconHash   string `json:"icon_hash"`
			Latitude   int    `json:"latitude"`
			LedOff     int    `json:"led_off"`
			Longitude  int    `json:"longitude"`
			Mac        string `json:"mac"`
			Model      string `json:"model"`
			OemID      string `json:"oemId"`
			OnTime     int    `json:"on_time"`
			RelayState int    `json:"relay_state"`
			Rssi       int    `json:"rssi"`
			SwVer      string `json:"sw_ver"`
			Type       string `json:"type"`
			Updating   int    `json:"updating"`
		} `json:"get_sysinfo"`
	} `json:"system"`
}

func systemInfo() (*systemInfoOutput, error) {
	res, err := plug.SystemInfo()
	if err != nil {
		log.Println("err:", err)
		return nil, err
	}

	var out systemInfoOutput
	err = json.Unmarshal([]byte(res), &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func main() {
	pinArg := flag.String("pin", "", "PIN used to pair the TP-Link HS1xx device with HomeKit")
	ipArg := flag.String("ip", "", "The IP address of the TP-Link HS1xx device")
	verboseArg := flag.Bool("v", false, "Whether or not hc log output is displayed")

	flag.Parse()

	if *verboseArg {
		hclog.Debug.Enable()
	}

	plug = hs1xxplug.Hs1xxPlug{IPAddress: *ipArg}
	info, err := systemInfo()
	if err != nil {
		log.Panic(err)
	}

	acc = accessory.NewSwitch(accessory.Info{
		Name:         info.System.Sysinfo.Alias,
		SerialNumber: info.System.Sysinfo.DeviceID,
		Manufacturer: "TP-Link",
		Model:        info.System.Sysinfo.Model,
	})
	config := hc.Config{Pin: *pinArg, IP: "10.0.12.241"}
	t, err := hc.NewIPTransport(config, acc.Accessory)
	if err != nil {
		log.Panic(err)
	}

	acc.Switch.On.SetValue(info.System.Sysinfo.RelayState == 1)

	acc.Switch.On.OnValueRemoteUpdate(func(on bool) {
		var err error
		if on {
			err = plug.TurnOn()
		} else {
			err = plug.TurnOff()
		}
		if err != nil {
			log.Println(err)
		}
	})

	hc.OnTermination(func() {
		t.Stop()
	})

	t.Start()
}
