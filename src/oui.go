package main

import (
	"bufio"
	"os"
	"strings"
)

// OUI Map
// Download oui.txt from
// http://standards-oui.ieee.org/oui/oui.txt

// OUIMap : OUI to Name Map
type OUIMap struct {
	Map map[string]string
}

// Open : Load OUI Data from file
func (oui *OUIMap) Open(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	oui.Map = make(map[string]string)
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := strings.TrimSpace(s.Text())
		if len(l) < 1 {
			continue
		}
		f := strings.Fields(l)
		if len(f) < 4 || f[1] != "(base" {
			continue
		}
		oui.Map[f[0]] = strings.Join(f[3:], " ")
	}
	return nil
}

// Find : Find  Vendor Name from MAC Address
func (oui *OUIMap) Find(mac string) string {
	mac = strings.TrimSpace(mac)
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	if len(mac) > 6 {
		mac = strings.ToUpper(mac)
		if n, ok := oui.Map[mac[:6]]; ok {
			return n
		}
	}
	return "Unknown"
}
