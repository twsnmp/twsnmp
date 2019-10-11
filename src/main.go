package main

import (
	"flag"
	"fmt"
	"time"
	"encoding/json"
	"context"
	"path/filepath"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
	mibdb "github.com/twsnmp/go-mibdb"
)

// Vars
var (
	AppName string
	BuiltAt   string
	dbPath    string
	debug     = flag.Bool("d", false, "enables the debug mode")
	startWindow     *astilectron.Window
	mainWindow     *astilectron.Window
	mib 	*mibdb.MIBDB
	dialogWindow     *astilectron.Window
	app        *astilectron.Astilectron
	aboutText = `TWSNMP Manager
Version 5.0.0
Copyright (c) 2019 Masayuki Yamai`
)

func main() {
	// Init
	flag.Parse()
	astilog.FlagInit()
	dbPath = flag.Arg(0)
	if dbPath != "" {
		if err := checkDB(dbPath); err != nil {
			astilog.Error(fmt.Sprintf("checkDB(Arg[0]) error=%v",err))
			dbPath = ""
		}
	}
	pctx := context.Background()
	ctx,cancel := context.WithCancel(pctx)	

	// Run bootstrap
	astilog.Debugf("Running app built at %s", BuiltAt)
	if err := bootstrap.Run(bootstrap.Options{
		Asset:    Asset,
		AssetDir: AssetDir,
		AstilectronOptions: astilectron.Options{
			AppName:            AppName,
			AppIconDarwinPath:  "resources/twsnmp.icns",
			AppIconDefaultPath: "resources/twsnmp.png",
		},
		Debug: *debug,
		MenuOptions: []*astilectron.MenuItemOptions{{
			Label: astilectron.PtrStr("File"),
			SubMenu: []*astilectron.MenuItemOptions{
				{
					Label: astilectron.PtrStr("TWSNMPについて"),
					OnClick: func(e astilectron.Event) (deleteListener bool) {
						if err := bootstrap.SendMessage(mainWindow, "about", aboutText, func(m *bootstrap.MessageIn) {
						}); err != nil {
							astilog.Error(fmt.Sprintf("sending about event failed err=%v", err))
						}
						return
					},
				},
				{
					Label: astilectron.PtrStr("TWSNMPを終了"),
					OnClick: func(e astilectron.Event) (deleteListener bool) {
						go func() {
							time.Sleep(time.Second*1)
							app.Stop()
						}()
						return
					},
				},
			},
		}},
		OnWait: func(a *astilectron.Astilectron, ws []*astilectron.Window, _ *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
			startWindow = ws[0]
			mainWindow = ws[1]
			dialogWindow = ws[2]
			app = a
			mibpath := filepath.Join(app.Paths().DataDirectory(), "resources","mib.txt")
			var err error
			mib, err = mibdb.NewMIBDB(mibpath)
			if err != nil {
				astilog.Fatalf("NewMIBDB failed err=%v", err)
			}
			startBackend(ctx)
			return nil
		},
		RestoreAssets: RestoreAssets,
		Windows: []*bootstrap.Window{
			{
				Homepage:       "start.html",
				MessageHandler: startMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:        astilectron.PtrBool(true),
					Modal:        astilectron.PtrBool(true),
					Show:          astilectron.PtrBool(true),
					Closable:          astilectron.PtrBool(false),
					Fullscreenable:          astilectron.PtrBool(false),
					Maximizable:          astilectron.PtrBool(false),
					Minimizable:          astilectron.PtrBool(false),
					Width:         astilectron.PtrInt(550),
					Height:        astilectron.PtrInt(550),
					TitleBarStyle: astilectron.TitleBarStyleHidden,
				},
			},
			{
				Homepage:       "main.html",
				MessageHandler: mainWindowMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:        astilectron.PtrBool(true),
					Show:          astilectron.PtrBool(false),
					Width:         astilectron.PtrInt(1024),
					Height:        astilectron.PtrInt(800),
					TitleBarStyle: astilectron.TitleBarStyleHidden,
				},
			},
			{
				Homepage:       "dialog.html",
				MessageHandler: dialogMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:        astilectron.PtrBool(true),
					Modal:        astilectron.PtrBool(true),
					Show:          astilectron.PtrBool(false),
					Fullscreenable:          astilectron.PtrBool(false),
					Maximizable:          astilectron.PtrBool(false),
					Minimizable:          astilectron.PtrBool(false),
					Width:         astilectron.PtrInt(500),
					Height:        astilectron.PtrInt(600),
					TitleBarStyle: astilectron.TitleBarStyleHidden,
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astilectron.PtrBool(true),
					},
				},
			},
		},
	}); err != nil {
		astilog.Fatal(fmt.Sprintf("running bootstrap failed err=%v", err))
	}
	stopDiscover()
	cancel()
	closeDB()
	astilog.Debug(fmt.Sprintf("End of main()"))
}

// startMessageHandler handles messages
func startMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
	case "exit":
		go func() {
			time.Sleep(time.Second*1)
			app.Stop()
		}()
	case "start":
		if len(m.Payload) > 0 {
			var fileName string
			if err := json.Unmarshal(m.Payload, &fileName); err != nil {
				astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
				return err.Error(),err
			}
			if err := checkDB(fileName); err != nil {
				astilog.Error(fmt.Sprintf("checkDB  error=%v", err))
				return err.Error(),err
			}
			dbPath = fileName
		}		
	}
	return "",nil
}

// Backen Process
func startBackend(ctx context.Context) {
	astilog.Debug("startBackend")
	go func() {
		if dbPath == "" {
			if err := bootstrap.SendMessage(startWindow, "selectDB",""); err != nil {
				astilog.Error(fmt.Sprintf("sendSendMessage selectDB error=%v", err))
			}
			for dbPath == "" {
				select {
				case <- ctx.Done():
					return
				case <- time.Tick(time.Second * 1):
					continue
				}
			}	
		} else {
			time.Sleep(time.Second * 2)
		}
		if err := openDB(dbPath);err != nil {
			astilog.Fatal(fmt.Sprintf("running bootstrap failed err=%v", err))
		}
		startWindow.Hide()
		mainWindow.Show()
		go mainWindowBackend(ctx)
		go eventLogger(ctx)
		addEventLog(eventLogEnt{
			Type: "system",
			Level:"info",
			Event: fmt.Sprintf("TWSNMP Manager Started. dbPath='%s'",dbPath),
		})
		go pollingBackend(ctx)
		go logger(ctx)
	}()
}