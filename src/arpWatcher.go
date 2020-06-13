package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var arpTable = make(map[string]string)
var localCheckAddrs []string

func arpWatcher(ctx context.Context) {
	astiLogger.Debug("start arpWacher")
	loadArpTableFromDB()
	checkArpTable()
	makeLoacalCheckAddrs()
	timer := time.NewTicker(time.Second * 300)
	pinger := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			astiLogger.Debug("Stop arpWatch")
			return
		case <-pinger.C:
			if len(localCheckAddrs) > 0 {
				a := localCheckAddrs[0]
				doPing(a, 1, 0, 64)
				localCheckAddrs[0] = ""
				localCheckAddrs = localCheckAddrs[1:]
			}
		case <-timer.C:
			checkArpTable()
			if len(localCheckAddrs) < 1 {
				makeLoacalCheckAddrs()
			} else {
				astiLogger.Infof("arpWatcher localCheckAddrs Count %d", len(localCheckAddrs))
			}
		}
	}
}

func makeLoacalCheckAddrs() {
	ifs, err := net.Interfaces()
	if err != nil {
		astiLogger.Errorf("makeLoacalCheckAddrs err=%v", err)
		return
	}
	localIPCount := 0
	localHitCount := 0
	for _, i := range ifs {
		if (i.Flags&net.FlagLoopback) == net.FlagLoopback ||
			(i.Flags&net.FlagUp) != net.FlagUp {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			cidr := a.String()
			ip, ipnet, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			if ip.To4() == nil {
				continue
			}
			astiLogger.Infof("arpWatch Check IP %s", cidr)
			for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
				if !ip.IsGlobalUnicast() || ip.IsMulticast() || ip.Equal(ip.Mask(ipnet.Mask)) {
					continue
				}
				localIPCount++
				sa := ip.String()
				if _, ok := arpTable[sa]; ok {
					localHitCount++
				}
				localCheckAddrs = append(localCheckAddrs, sa)
			}
		}
	}
	lau := 0.0
	if localIPCount > 0 {
		lau = 100.0 * float64(localHitCount) / float64(localIPCount)
	}
	addEventLog(eventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf("ARP監視 ローカルアドレス使用量 %d/%d %.2f%%", localHitCount, localIPCount, lau),
	})
	astiLogger.Infof("Local Address Usage %d/%d", localHitCount, localIPCount)
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func checkArpTable() {
	if runtime.GOOS == "windows" {
		checkArpTableWindows()
	} else {
		checkArpTableUnix()
	}
	checkNodeMAC()
}

func checkArpTableWindows() {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		astiLogger.Errorf("checkArpTable err=%v", err)
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		updateArpTable(fields[0], fields[1])
	}
}

func checkArpTableUnix() {
	out, err := exec.Command("arp", "-an").Output()
	if err != nil {
		astiLogger.Errorf("checkArpTable err=%v", err)
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// strip brackets around IP
		ip := strings.Replace(fields[1], "(", "", -1)
		ip = strings.Replace(ip, ")", "", -1)
		updateArpTable(ip, fields[3])
	}
}

func updateArpTable(ip, mac string) {
	if !strings.Contains(ip, ".") || !strings.ContainsAny(mac, ":-") {
		return
	}
	mac = normMACAddr(mac)
	if strings.HasPrefix(mac, "FF") || strings.HasPrefix(mac, "01") {
		return
	}
	deviceReportCh <- &deviceReportEnt{
		IP:   ip,
		MAC:  mac,
		Time: time.Now().UnixNano(),
	}
	m, ok := arpTable[ip]
	if !ok {
		// New
		updateArpEnt(ip, mac)
		logCh <- &logEnt{
			Time: time.Now().UnixNano(),
			Type: "arplog",
			Log:  fmt.Sprintf("New,%s,%s", ip, mac),
		}
		astiLogger.Infof("New %s %s", ip, mac)
		return
	}
	if mac != m {
		// Change
		updateArpEnt(ip, mac)
		logCh <- &logEnt{
			Time: time.Now().UnixNano(),
			Type: "arplog",
			Log:  fmt.Sprintf("Change,%s,%s,%s", ip, m, mac),
		}
		astiLogger.Infof("Change %s %s -> %s", ip, m, mac)
		return
	}
	// No Change
}

func normMACAddr(m string) string {
	m = strings.Replace(m, "-", ":", -1)
	a := strings.Split(m, ":")
	r := ""
	for _, e := range a {
		if r != "" {
			r += ":"
		}
		if len(e) == 1 {
			r += "0"
		}
		r += e
	}
	return strings.ToUpper(r)
}

// ノードリストのMACアドレスをチェックする
func checkNodeMAC() {
	for _, n := range nodes {
		if m, ok := arpTable[n.IP]; ok {
			if !strings.Contains(n.MAC, m) {
				new := m
				v := oui.Find(m)
				if v != "" {
					new += fmt.Sprintf("(%s)", v)
				}
				addEventLog(eventLogEnt{
					Type:     "arpwatch",
					Level:    mapConf.ArpWatchLevel,
					NodeID:   n.ID,
					NodeName: n.Name,
					Event:    fmt.Sprintf("MACアドレス変化 %s -> %s", n.MAC, new),
				})
				n.MAC = new
				updateNode(n)
			}
		}
	}
}
