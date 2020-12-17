package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func notifyBackend(ctx context.Context) {
	lastSendReport := time.Now().Add(time.Hour * time.Duration(-24))
	lastLog := ""
	lastLog = checkNotify(lastLog)
	timer := time.NewTicker(time.Second * 60)
	i := 0
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			i++
			if i >= notifyConf.Interval {
				i = 0
				lastLog = checkNotify(lastLog)
			}
			checkExecCmd()
			if notifyConf.Report == "send" && lastSendReport.Day() != time.Now().Day() {
				lastSendReport = time.Now()
				sendReport()
			}
		}
	}
}

func getLevelNum(l string) int {
	switch l {
	case "high":
		return 0
	case "low":
		return 1
	case "warn":
		return 2
	}
	return 3
}

var lastExecLevel = -1

func checkExecCmd() {
	if notifyConf.ExecCmd == "" {
		return
	}
	execLevel := 3
	for _, n := range nodes {
		ns := getLevelNum(n.State)
		if execLevel > ns {
			execLevel = ns
			if ns == 0 {
				break
			}
		}
	}
	if execLevel != lastExecLevel {
		err := execNotifyCmd(execLevel)
		r := ""
		if err != nil {
			astiLogger.Errorf("execNotifyCmd err=%v", err)
			r = fmt.Sprintf("エラー=%v", err)
		}
		addEventLog(eventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("外部通知コマンド実行 レベル=%d %s", execLevel, r),
		})
		lastExecLevel = execLevel
	}
}

func execNotifyCmd(level int) error {
	cl := strings.Split(notifyConf.ExecCmd, " ")
	if len(cl) < 1 {
		return nil
	}
	strLevel := fmt.Sprintf("%d", level)
	if len(cl) == 1 {
		return exec.Command(cl[0]).Start()
	}
	for i, v := range cl {
		if v == "$level" {
			cl[i] = strLevel
		}
	}
	return exec.Command(cl[0], cl[1:]...).Start()
}

func checkNotify(lastLog string) string {
	list := getEventLogList(lastLog, 1000)
	if len(list) > 0 {
		nl := getLevelNum(notifyConf.Level)
		if nl == 3 {
			return fmt.Sprintf("%016x", list[0].Time)
		}
		body := []string{}
		repair := []string{}
		ti := time.Now().Add(time.Duration(-notifyConf.Interval) * time.Minute).UnixNano()
		for _, l := range list {
			if ti > l.Time {
				continue
			}
			if notifyConf.NotifyRepair && l.Level == "repair" {
				a := strings.Split(l.Event, ":")
				if len(a) < 5 {
					continue
				}
				// 復帰前の状態を確認する
				n := getLevelNum(a[2])
				if n > nl {
					continue
				}
				ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
				repair = append(repair, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
				continue
			}
			n := getLevelNum(l.Level)
			if n > nl {
				continue
			}
			ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
			body = append(body, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
		}
		if len(body) > 0 {
			err := sendMail(notifyConf.Subject, strings.Join(body, "\r\n"))
			r := ""
			if err != nil {
				astiLogger.Errorf("sendMail err=%v", err)
				r = fmt.Sprintf("失敗 エラー=%v", err)
			}
			addEventLog(eventLogEnt{
				Type:  "system",
				Level: "info",
				Event: fmt.Sprintf("通知メール送信 %s", r),
			})
		}
		if len(repair) > 0 {
			err := sendMail(notifyConf.Subject+"(復帰)", strings.Join(repair, "\r\n"))
			r := ""
			if err != nil {
				astiLogger.Errorf("sendMail err=%v", err)
				r = fmt.Sprintf("失敗 エラー=%v", err)
			}
			addEventLog(eventLogEnt{
				Type:  "system",
				Level: "info",
				Event: fmt.Sprintf("復帰通知メール送信 %s", r),
			})
		}
		lastLog = fmt.Sprintf("%016x", list[0].Time)
	}
	return lastLog
}

func sendMail(subject, body string) error {
	if notifyConf.MailServer == "" || notifyConf.MailFrom == "" || notifyConf.MailTo == "" {
		return nil
	}
	tlsconfig := &tls.Config{
		ServerName:         notifyConf.MailServer,
		InsecureSkipVerify: notifyConf.InsecureSkipVerify,
	}
	c, err := smtp.Dial(notifyConf.MailServer)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.StartTLS(tlsconfig); err != nil {
		astiLogger.Warnf("StartTLS err=%s", err)
	}
	msv := notifyConf.MailServer
	a := strings.SplitN(notifyConf.MailServer, ":", 2)
	if len(a) == 2 {
		msv = a[0]
	}
	if notifyConf.User != "" {
		auth := smtp.PlainAuth("", notifyConf.User, notifyConf.Password, msv)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}
	if err = c.Mail(notifyConf.MailFrom); err != nil {
		return err
	}
	for _, rcpt := range strings.Split(notifyConf.MailTo, ",") {
		if err = c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()
	body = convNewline(body, "\r\n")
	message := makeMailMessage(notifyConf.MailFrom, notifyConf.MailTo, subject, body)
	_, _ = w.Write([]byte(message))
	_ = c.Quit()
	astiLogger.Infof("Send Mail to %s", notifyConf.MailTo)
	return nil
}

func convNewline(str, nlcode string) string {
	return strings.NewReplacer(
		"\r\n", nlcode,
		"\r", nlcode,
		"\n", nlcode,
	).Replace(str)
}

func sendTestMail(testConf *notifyConfEnt) error {
	tlsconfig := &tls.Config{
		ServerName:         testConf.MailServer,
		InsecureSkipVerify: testConf.InsecureSkipVerify,
	}
	c, err := smtp.Dial(testConf.MailServer)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.StartTLS(tlsconfig); err != nil {
		astiLogger.Warnf("StartTLS err=%s", err)
	}
	msv := testConf.MailServer
	a := strings.SplitN(testConf.MailServer, ":", 2)
	if len(a) == 2 {
		msv = a[0]
	}
	if testConf.User != "" {
		auth := smtp.PlainAuth("", testConf.User, testConf.Password, msv)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}
	if err = c.Mail(testConf.MailFrom); err != nil {
		return err
	}
	for _, rcpt := range strings.Split(testConf.MailTo, ",") {
		if err = c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()
	body := "Test Mail.\r\n試験メール.\r\n"
	message := makeMailMessage(testConf.MailFrom, testConf.MailTo, testConf.Subject, body)
	_, _ = w.Write([]byte(message))
	_ = c.Quit()
	return nil
}

func makeMailMessage(from, to, subject, body string) string {
	var header bytes.Buffer
	header.WriteString("From: " + from + "\r\n")
	header.WriteString("To: " + to + "\r\n")
	header.WriteString(encodeSubject(subject))
	header.WriteString("MIME-Version: 1.0\r\n")
	header.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	header.WriteString("Content-Transfer-Encoding: base64\r\n")

	var message bytes.Buffer = header
	message.WriteString("\r\n")
	message.WriteString(add76crlf(base64.StdEncoding.EncodeToString([]byte(body))))

	return message.String()
}

// 76バイト毎にCRLFを挿入する
func add76crlf(msg string) string {
	var buffer bytes.Buffer
	for k, c := range strings.Split(msg, "") {
		buffer.WriteString(c)
		if k%76 == 75 {
			buffer.WriteString("\r\n")
		}
	}
	return buffer.String()
}

// UTF8文字列を指定文字数で分割
func utf8Split(utf8string string, length int) []string {
	resultString := []string{}
	var buffer bytes.Buffer
	for k, c := range strings.Split(utf8string, "") {
		buffer.WriteString(c)
		if k%length == length-1 {
			resultString = append(resultString, buffer.String())
			buffer.Reset()
		}
	}
	if buffer.Len() > 0 {
		resultString = append(resultString, buffer.String())
	}
	return resultString
}

// サブジェクトをMIMEエンコードする
func encodeSubject(subject string) string {
	var buffer bytes.Buffer
	buffer.WriteString("Subject:")
	for _, line := range utf8Split(subject, 13) {
		buffer.WriteString(" =?utf-8?B?")
		buffer.WriteString(base64.StdEncoding.EncodeToString([]byte(line)))
		buffer.WriteString("?=\r\n")
	}
	return buffer.String()
}

func sendFeedback(msg string) error {
	msg += fmt.Sprintf("\n-----\n%s:%s\n", runtime.GOOS, runtime.GOARCH)
	values := url.Values{}
	values.Set("msg", msg)
	values.Add("hash", calcHash(msg))

	req, err := http.NewRequest(
		"POST",
		"https://lhx98.linkclub.jp/twise.co.jp/cgi-bin/twsnmpfb.cgi",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		astiLogger.Errorf("sendFeedback  err=%v", err)
		return err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		astiLogger.Errorf("sendFeedback  err=%v", err)
		return err
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		astiLogger.Errorf("sendFeedback  err=%v", err)
		return err
	}
	if string(r) != "OK" {
		return fmt.Errorf("resp is '%s'", r)
	}
	return nil
}

func calcHash(msg string) string {
	h := sha256.New()
	if _, err := h.Write([]byte(msg + time.Now().Format("2006/01/02T15"))); err != nil {
		astiLogger.Errorf("calcHash  err=%v", err)
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func sendReport() {
	body := []string{}
	logs := []string{}
	body = append(body, "【現在のマップ情報】")
	body = append(body, getMapInfo()...)
	body = append(body, "")
	list := getEventLogList("", 5000)
	high := 0
	low := 0
	warn := 0
	normal := 0
	other := 0
	if len(list) > 0 {
		ti := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
		for _, l := range list {
			if ti > l.Time {
				continue
			}
			switch l.Level {
			case "high":
				high++
			case "low":
				low++
			case "warn":
				warn++
			case "normal", "repair":
				normal++
			default:
				other++
			}
			ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
			logs = append(logs, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
		}
	}
	body = append(body, "【48時間以内に新しく発見したデバイス】")
	body = append(body, getNewDevice()...)
	body = append(body, "")
	body = append(body, "【48時間以内に新しく発見したユーザーID】")
	body = append(body, getNewUser()...)
	body = append(body, "")
	body = append(body, "【24時間以内の状態別ログ件数】")
	body = append(body, fmt.Sprintf("重度=%d,軽度=%d,注意=%d,正常=%d,その他=%d", high, low, warn, normal, other))
	body = append(body, "")
	body = append(body, "【最新24時間のログ】")
	body = append(body, logs...)
	if err := sendMail(fmt.Sprintf("TWSNMP定期レポート %s", time.Now().Format(time.RFC3339)), strings.Join(body, "\r\n")); err != nil {
		astiLogger.Errorf("sendMail err=%v", err)
	} else {
		addEventLog(eventLogEnt{
			Type:  "system",
			Level: "info",
			Event: "定期レポートメール送信",
		})
	}
}

func getMapInfo() []string {
	high := 0
	low := 0
	warn := 0
	normal := 0
	repair := 0
	unknown := 0
	for _, n := range nodes {
		switch n.State {
		case "high":
			high++
		case "low":
			low++
		case "warn":
			warn++
		case "normal":
			normal++
		case "repair":
			repair++
		default:
			unknown++
		}
	}
	state := "unknown"
	if high > 0 {
		state = "high"
	} else if low > 0 {
		state = "low"
	} else if warn > 0 {
		state = "warn"
	} else if normal+repair > 0 {
		state = "normal"
	}
	return []string{
		fmt.Sprintf("MAP状態=%s", state),
		fmt.Sprintf("重度=%d,軽度=%d,注意=%d,復帰=%d,正常=%d,不明=%d", high, low, warn, repair, normal, unknown),
		fmt.Sprintf("データベースサイズ=%s", dbStats.Size),
	}
}

func getNewDevice() []string {
	st := time.Now().Add(time.Duration(-48) * time.Hour).UnixNano()
	ret := []string{}
	for _, d := range devices {
		if d.FirstTime < st {
			continue
		}
		ret = append(ret, fmt.Sprintf("%s,%s,%s,%s", d.Name, d.IP, d.ID, d.Vendor))
	}
	return (ret)
}

func getNewUser() []string {
	st := time.Now().Add(time.Duration(-48) * time.Hour).UnixNano()
	ret := []string{}
	for _, u := range users {
		if u.FirstTime < st {
			continue
		}
		ret = append(ret, fmt.Sprintf("%s,%s,%s", u.UserID, u.ServerName, u.Server))
	}
	return (ret)
}
