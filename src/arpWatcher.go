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
var macToIPTable = make(map[string]string)
var localCheckAddrs []string

func arpWatcher(ctx context.Context) {
	astiLogger.Debug("start arpWacher")
	if err := loadArpTableFromDB(); err != nil {
		astiLogger.Errorf("loadArpTableFromDB err=%v", err)
	}
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

var lastLocalAddressUsage = 0.0

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
			(i.Flags&net.FlagUp) != net.FlagUp ||
			(i.Flags&net.FlagPointToPoint) == net.FlagPointToPoint ||
			len(i.HardwareAddr) != 6 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		ipMap := make(map[string]bool)
		for _, a := range addrs {
			cidr := a.String()
			ipTmp, ipnet, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			ip := ipTmp.To4()
			if ip == nil {
				continue
			}
			mask := ipnet.Mask
			broadcast := net.IP(make([]byte, 4))
			for i := range ip {
				broadcast[i] = ip[i] | ^mask[i]
			}
			astiLogger.Infof("arpWatch Check IP %s %s", cidr, broadcast)
			for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
				if !ip.IsGlobalUnicast() || ip.IsMulticast() ||
					ip.Equal(ip.Mask(ipnet.Mask)) || ip.Equal(broadcast) {
					continue
				}
				sa := ip.String()
				if _, ok := ipMap[sa]; ok {
					// 同じIPアドレスを追加しない。
					continue
				}
				ipMap[sa] = true
				localIPCount++
				if _, ok := arpTable[sa]; ok {
					localHitCount++
					continue
				}
				localCheckAddrs = append(localCheckAddrs, sa)
			}
		}
	}
	lau := 0.0
	if localIPCount > 0 {
		lau = 100.0 * float64(localHitCount) / float64(localIPCount)
	}
	if lastLocalAddressUsage == lau {
		return
	}
	lastLocalAddressUsage = lau
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
		if err := updateArpEnt(ip, mac); err != nil {
			astiLogger.Errorf("updateArpEnt err=%v", err)
		}
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
		if err := updateArpEnt(ip, mac); err != nil {
			astiLogger.Errorf("updateArpEnt err=%v", err)
		}
		logCh <- &logEnt{
			Time: time.Now().UnixNano(),
			Type: "arplog",
			Log:  fmt.Sprintf("Change,%s,%s,%s", ip, m, mac),
		}
		astiLogger.Infof("Change %s %s -> %s", ip, m, mac)
		return
	}
	// No Change
	// Check MAC to IP Table
	ipot, ok := macToIPTable[mac]
	if !ok {
		astiLogger.Infof("New MAC to IP %s %s", mac, ip)
		macToIPTable[mac] = ip
	} else if ip != ipot {
		astiLogger.Infof("Change MAC to IP %s %s -> %s", mac, ipot, ip)
		macToIPTable[mac] = ip
	}
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
		if n.AddrMode == "mac" {
			checkFixMACMode(n)
			continue
		}
		if n.AddrMode == "host" {
			checkFixHostMode(n)
			continue
		}
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
				if err := updateNode(n); err != nil {
					astiLogger.Errorf("updateNode err=%v", err)
				}
			}
		}
	}
}

func checkFixMACMode(n *nodeEnt) {
	if n.MAC == "" {
		if mac, ok := arpTable[n.IP]; ok {
			v := oui.Find(mac)
			if v != "" {
				mac += fmt.Sprintf("(%s)", v)
			}
			n.MAC = mac
			if err := updateNode(n); err != nil {
				astiLogger.Errorf("checkFixMACMode err=%v", err)
			} else {
				addEventLog(eventLogEnt{
					Type:     "system",
					Level:    "info",
					NodeID:   n.ID,
					NodeName: n.Name,
					Event:    fmt.Sprintf("MACアドレス固定ノードのアドレス取得 %s", n.MAC),
				})
			}
		}
		return
	}
	a := strings.Split(n.MAC, "(")
	if len(a) < 1 {
		return
	}
	key := strings.TrimSpace(a[0])
	if ip, ok := macToIPTable[key]; ok {
		if ip != n.IP {
			oldIP := n.IP
			n.IP = ip
			if err := updateNode(n); err != nil {
				astiLogger.Errorf("checkFixMACMode err=%v", err)
			} else {
				addEventLog(eventLogEnt{
					Type:     "system",
					Level:    "info",
					NodeID:   n.ID,
					NodeName: n.Name,
					Event:    fmt.Sprintf("MACアドレス固定ノード'%s'のIPアドレスが'%s'から'%s'に変化", n.MAC, oldIP, ip),
				})
			}
		}
	}
}

func checkFixHostMode(n *nodeEnt) {
	ips, err := net.LookupHost(n.Name)
	if err != nil {
		astiLogger.Errorf("checkFixHostMode err=%v", err)
		return
	}
	hitIP := ""
	for _, ip := range ips {
		if n.IP == ip {
			return
		}
		if strings.Contains(ip, ":") || hitIP != "" {
			continue
		}
		nIP := net.ParseIP(ip)
		if nIP.IsGlobalUnicast() {
			hitIP = ip
		}
	}
	if hitIP == "" {
		astiLogger.Errorf("checkFixHostMode no ip found")
		return
	}
	oldIP := n.IP
	n.IP = hitIP
	if err := updateNode(n); err != nil {
		astiLogger.Errorf("checkFixMACMode err=%v", err)
	} else {
		addEventLog(eventLogEnt{
			Type:     "system",
			Level:    "info",
			NodeID:   n.ID,
			NodeName: n.Name,
			Event:    fmt.Sprintf("ホスト名固定ノード'%s'のIPアドレスが'%s'から''%sに変化", n.Name, oldIP, hitIP),
		})
	}
}
