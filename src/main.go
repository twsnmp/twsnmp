package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
	mibdb "github.com/twsnmp/go-mibdb"
)

// Vars
var (
	AppName       string
	BuiltAt       string
	dbPath        string
	debug         = flag.Bool("d", false, "enables the debug mode")
	startWindow   *astilectron.Window
	mainWindow    *astilectron.Window
	nodeWindow    *astilectron.Window
	logWindow     *astilectron.Window
	pollingWindow *astilectron.Window
	mibWindow     *astilectron.Window
	mib           *mibdb.MIBDB
	app           *astilectron.Astilectron
	aboutText     = `TWSNMP Manager
Version 5.0.0
Copyright (c) 2019 Masayuki Yamai`
)

// Define errors
var (
	errNoPayload     = fmt.Errorf("No Payload")
	errInvalidNode   = fmt.Errorf("Invalid Node")
	errInvalidParams = fmt.Errorf("Invald Params")
	errDBNotOpen     = fmt.Errorf("DB Not Open")
	errInvalidID     = fmt.Errorf("Invalid ID")
)

func main() {
	// Init
	flag.Parse()
	astilog.FlagInit()
	dbPath = flag.Arg(0)
	if dbPath != "" {
		if err := checkDB(dbPath); err != nil {
			astilog.Error(fmt.Sprintf("checkDB(Arg[0]) error=%v", err))
			dbPath = ""
		}
	}
	pctx := context.Background()
	ctx, cancel := context.WithCancel(pctx)

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
							time.Sleep(time.Second * 1)
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
			nodeWindow = ws[2]
			logWindow = ws[3]
			pollingWindow = ws[4]
			mibWindow = ws[5]
			app = a
			mibpath := filepath.Join(app.Paths().DataDirectory(), "resources", "mib.txt")
			var err error
			mib, err = mibdb.NewMIBDB(mibpath)
			if err != nil {
				astilog.Fatalf("NewMIBDB failed err=%v", err)
			}
			startBackend(ctx)
			mainWindow.On(astilectron.EventNameWindowEventClosed, func(e astilectron.Event) (deleteListener bool) {
				astilog.Debug("Main Window Closed")
				app.Stop()
				return
			})
			return nil
		},
		RestoreAssets: RestoreAssets,
		Windows: []*bootstrap.Window{
			{
				Homepage:       "start.html",
				MessageHandler: startMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astilectron.PtrBool(true),
					Modal:          astilectron.PtrBool(true),
					Show:           astilectron.PtrBool(true),
					Closable:       astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(false),
					Width:          astilectron.PtrInt(450),
					Height:         astilectron.PtrInt(500),
					TitleBarStyle:  astilectron.TitleBarStyleHidden,
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
				Homepage:       "node.html",
				MessageHandler: nodeMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astilectron.PtrBool(true),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(800),
					Height:         astilectron.PtrInt(500),
					TitleBarStyle:  astilectron.TitleBarStyleHidden,
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astilectron.PtrBool(true),
					},
				},
			},
			{
				Homepage:       "log.html",
				MessageHandler: logMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astilectron.PtrBool(true),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(1000),
					Height:         astilectron.PtrInt(700),
					TitleBarStyle:  astilectron.TitleBarStyleHidden,
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astilectron.PtrBool(true),
					},
				},
			},
			{
				Homepage:       "polling.html",
				MessageHandler: pollingMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astilectron.PtrBool(true),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(900),
					Height:         astilectron.PtrInt(750),
					TitleBarStyle:  astilectron.TitleBarStyleHidden,
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astilectron.PtrBool(true),
					},
				},
			},
			{
				Homepage:       "mib.html",
				MessageHandler: mibMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astilectron.PtrBool(true),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(800),
					Height:         astilectron.PtrInt(500),
					TitleBarStyle:  astilectron.TitleBarStyleHidden,
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
func startMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "exit":
		go func() {
			time.Sleep(time.Second * 1)
			app.Stop()
		}()
	case "start":
		if len(m.Payload) > 0 {
			var fileName string
			if err := json.Unmarshal(m.Payload, &fileName); err != nil {
				astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
				return err.Error(), err
			}
			if err := checkDB(fileName); err != nil {
				astilog.Error(fmt.Sprintf("checkDB  error=%v", err))
				return err.Error(), err
			}
			dbPath = fileName
		}
	}
	return "", nil
}

// Backen Process
func startBackend(ctx context.Context) {
	astilog.Debug("startBackend")
	go func() {
		if dbPath == "" {
			if err := bootstrap.SendMessage(startWindow, "selectDB", ""); err != nil {
				astilog.Error(fmt.Sprintf("sendSendMessage selectDB error=%v", err))
			}
			for dbPath == "" {
				select {
				case <-ctx.Done():
					return
				case <-time.Tick(time.Second * 1):
					continue
				}
			}
		} else {
			time.Sleep(time.Second * 2)
		}
		if err := openDB(dbPath); err != nil {
			astilog.Fatal(fmt.Sprintf("running bootstrap failed err=%v", err))
		}
		go mainWindowBackend(ctx)
		go eventLogger(ctx)
		addEventLog(eventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("TWSNMP起動 データベース='%s'", dbPath),
		})
		go pollingBackend(ctx)
		go logger(ctx)
		startWindow.Hide()
		mainWindow.Show()
	}()
}
