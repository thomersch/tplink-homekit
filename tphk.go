package main

import (
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
)

const iv = 171

type sysInfo struct {
	System struct {
		GetSysinfo struct {
			SWVer    string `json:"sw_ver"`
			DeviceID string `json:"deviceId"`
			Alias    string `json:"alias"`
		} `json:"get_sysinfo"`
	} `json:"system"`
}

type device struct {
	Addr     string
	DevID    string
	Alias    string
	Firmware string
}

func (d *device) Init() error {
	return d.info()
}

func (d *device) dial() (net.Conn, error) {
	var err error
	sock, err := net.Dial("tcp", d.Addr)
	if err != nil {
		return nil, err
	}
	return sock, nil
}

func (d *device) info() error {
	var si sysInfo
	buf, err := d.do([]byte(`{"system":{"get_sysinfo":{}}}`))
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, &si)
	if err != nil {
		return err
	}
	d.DevID = si.System.GetSysinfo.DeviceID
	d.Alias = si.System.GetSysinfo.Alias
	d.Firmware = strings.Split(si.System.GetSysinfo.SWVer, " ")[0]
	return nil
}

func (d *device) Announce() *accessory.Accessory {
	i := accessory.Info{
		Name:         d.Alias,
		SerialNumber: d.DevID,
		Manufacturer: "TP-Link",
	}
	acc := accessory.NewSwitch(i)
	acc.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on {
			d.On()
		} else {
			d.Off()
		}
	})
	return acc.Accessory
}

func (d *device) On() {
	res, err := d.do([]byte(`{"system":{"set_relay_state":{"state":1}}}`))
	if err != nil {
		log.Println(res, err)
	}
}

func (d *device) Off() {
	res, err := d.do([]byte(`{"system":{"set_relay_state":{"state":0}}}`))
	if err != nil {
		log.Println(res, err)
	}
}

func (d *device) do(cmd []byte) ([]byte, error) {
	conn, err := d.dial()
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(encrypt(cmd))
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	buf = decrypt(buf)
	return buf, nil
}

func encrypt(buf []byte) []byte {
	var (
		key = byte(iv)
		out = make([]byte, 4)
	)
	binary.BigEndian.PutUint32(out[0:], uint32(len(buf)))
	for _, b := range buf {
		a := key ^ b
		key = a
		out = append(out, a)
	}
	return out
}

func decrypt(buf []byte) []byte {
	var (
		key = byte(iv)
		out []byte
	)
	if len(buf) < 4 {
		return nil
	}
	buf = buf[4:]
	for _, b := range buf {
		a := key ^ b
		key = b
		out = append(out, a)
	}
	return out
}

func main() {
	devStr := os.Getenv("TPLINK_OUTLET_HOSTS")
	devNames := strings.Split(devStr, ",")

	log.Printf("Going to look for: %v", devNames)

	var accs []*accessory.Accessory

	for _, dev := range devNames {
		d := device{Addr: dev + ":9999"}
		err := d.Init()
		if err != nil {
			continue
		}
		accs = append(accs, d.Announce())
	}

	config := hc.Config{Pin: "19011011"}
	t, err := hc.NewIPTransport(config, accessory.New(accessory.Info{Name: "TPLinkBridge", SerialNumber: "1", Manufacturer: "Nobody", Model: "Default"}, accessory.TypeBridge), accs...)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})
	t.Start()
}
