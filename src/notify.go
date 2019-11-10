package main

import (
	"context"
	"crypto/tls"
	"net/smtp"
	"time"
	"fmt"
	"strings"
	astilog "github.com/asticode/go-astilog"
)

func notifyBackend(ctx context.Context) {
	lastLog := fmt.Sprintf("%016x", time.Now().Add(time.Minute * time.Duration(-notifyConf.Interval)).UnixNano())
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(time.Second * 60):
			i++
			if i > notifyConf.Interval {
				i = 0
				lastLog = checkSendMail(lastLog)
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

func checkSendMail(lastLog string) string{
	list := getEventLogList(lastLog, 1000)
	if len(list) > 0 {
		nl := getLevelNum(notifyConf.Level)
		if nl == 3 {
			return fmt.Sprintf("%016x", list[0].Time)
		}
		body := []string{}
		for _,l := range list {
			n := getLevelNum(l.Level)
			if n > nl {
				continue
			}
			ts := time.Unix(0,l.Time).Local().Format(time.RFC3339Nano)
			body = append(body,fmt.Sprintf("%s,%s,%s,%s,%s",l.Level,ts,l.Type,l.NodeName,l.Event))
		}
		if len(body) > 0 {
			if err := sendMail(notifyConf.Subject,strings.Join(body,"\r\n"));err != nil {
				astilog.Errorf("sendMail err=%v",err)
			}
		}
		return fmt.Sprintf("%016x", list[0].Time)
	}
	return lastLog
}

func sendMail(subject,body string) error {
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
		return err
	}
	if notifyConf.User != "" {
		auth := smtp.PlainAuth("", notifyConf.User,notifyConf.Password,notifyConf.MailServer)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}
	if err = c.Mail(notifyConf.MailFrom); err != nil {
		return err
	}
	for _,rcpt := range strings.Split(notifyConf.MailTo,",") {
		if err = c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()
	message := ""
	message += "From: " + notifyConf.MailFrom + "\r\n"
	message += "To: " +  notifyConf.MailTo + "\r\n"
	message += "Subject: " +  subject + "\r\n"
	message += "\r\n" + convNewline(body,"\r\n")
	w.Write([]byte(message))
	c.Quit()
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
		return err
	}
	if testConf.User != "" {
		auth := smtp.PlainAuth("", testConf.User,testConf.Password,testConf.MailServer)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}
	if err = c.Mail(testConf.MailFrom); err != nil {
		return err
	}
	for _,rcpt := range strings.Split(testConf.MailTo,",") {
		if err = c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()
	message := ""
	message += "From: " + testConf.MailFrom + "\r\n"
	message += "To: " +  testConf.MailTo + "\r\n"
	message += "Subject: " +  testConf.Subject + "\r\n"
	message += "\r\n Test Mail.\r\n"
	w.Write([]byte(message))
	c.Quit()
	return nil
}