package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
	mibdb "github.com/twsnmp/go-mibdb"
)

// Vars
var (
	AppName           string
	BuiltAt           string
	dbPath            string
	debug             = flag.Bool("d", false, "enables the debug mode")
	cpuprofile        = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile        = flag.String("memprofile", "", "write memory profile to `file`")
	startWindow       *astilectron.Window
	mainWindow        *astilectron.Window
	nodeWindow        *astilectron.Window
	pollingListWindow *astilectron.Window
	logWindow         *astilectron.Window
	pollingWindow     *astilectron.Window
	mibWindow         *astilectron.Window
	aiWindow          *astilectron.Window
	mib               *mibdb.MIBDB
	oui               = &OUIMap{}
	app               *astilectron.Astilectron
	aboutText         = `TWSNMP Manager
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
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			astilog.Fatalf("could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			astilog.Fatalf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			astilog.Fatalf("could not create memory profile: %v", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			astilog.Fatalf("could not write memory profile:%v", err)
		}
	}
	logConf := astilog.FlagConfig()
	logConf.FullTimestamp = true
	logConf.DisableTimestamp = false
	astilog.SetLogger(astilog.New(logConf))
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
		}, {
			Label: astilectron.PtrStr("Window"),
			SubMenu: []*astilectron.MenuItemOptions{
				{
					Checked: astilectron.PtrBool(true), Label: astilectron.PtrStr("マップ"),
					Type: astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(mainWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astilectron.PtrStr("ノード情報"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(nodeWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astilectron.PtrStr("ポーリングリスト"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(pollingListWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astilectron.PtrStr("ログ表示"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(logWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astilectron.PtrStr("ポーリング"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(pollingWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astilectron.PtrStr("MIBブラウザー"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(mibWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astilectron.PtrStr("AI分析"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(aiWindow, *e.MenuItemOptions.Checked)
						if *debug {
							if *e.MenuItemOptions.Checked {
								aiWindow.OpenDevTools()
							} else {
								aiWindow.CloseDevTools()
							}
						}
						return false
					},
				},
			},
		}},
		OnWait: func(a *astilectron.Astilectron, w []*astilectron.Window, m *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
			startWindow = w[0]
			mainWindow = w[1]
			nodeWindow = w[2]
			pollingListWindow = w[3]
			logWindow = w[4]
			pollingWindow = w[5]
			mibWindow = w[6]
			aiWindow = w[7]
			app = a
			path := filepath.Join(app.Paths().DataDirectory(), "resources", "mib.txt")
			var err error
			mib, err = mibdb.NewMIBDB(path)
			if err != nil {
				astilog.Fatalf("NewMIBDB failed err=%v", err)
			}
			path = filepath.Join(app.Paths().DataDirectory(), "resources", "tlsparams.csv")
			loadTLSParamsMap(path)
			path = filepath.Join(app.Paths().DataDirectory(), "resources", "oui.txt")
			oui.Open(path)
			startBackend(ctx)
			mainWindow.On(astilectron.EventNameWindowEventClosed, func(e astilectron.Event) (deleteListener bool) {
				astilog.Debug("Main Window Closed")
				app.Stop()
				return
			})
			for i, w := range w {
				if i < 2 {
					continue
				}
				mi, err := m.Item(1, i-1)
				if err != nil {
					continue
				}
				if w != aiWindow {
					mi.SetVisible(false)
				}
				w.On(astilectron.EventNameWindowEventHide, func(e astilectron.Event) (deleteListener bool) {
					mi.SetChecked(false)
					return
				})
				w.On(astilectron.EventNameWindowEventMinimize, func(e astilectron.Event) (deleteListener bool) {
					mi.SetChecked(false)
					return
				})
				w.On(astilectron.EventNameWindowEventShow, func(e astilectron.Event) (deleteListener bool) {
					mi.SetChecked(true)
					mi.SetVisible(true)
					return
				})
			}
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
					Frame:          astilectron.PtrBool(false),
					Closable:       astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(false),
					Width:          astilectron.PtrInt(450),
					Height:         astilectron.PtrInt(500),
				},
			},
			{
				Homepage:       "main.html",
				MessageHandler: mainWindowMessageHandler,
				Options: &astilectron.WindowOptions{
					Center: astilectron.PtrBool(true),
					Show:   astilectron.PtrBool(false),
					Width:  astilectron.PtrInt(1024),
					Height: astilectron.PtrInt(800),
				},
			},
			{
				Homepage:       "node.html",
				MessageHandler: nodeMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astilectron.PtrBool(true),
					Frame:          astilectron.PtrBool(false),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(850),
					Height:         astilectron.PtrInt(450),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astilectron.PtrBool(true),
					},
				},
			},
			{
				Homepage:       "pollingList.html",
				MessageHandler: pollingListMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astilectron.PtrBool(true),
					Frame:          astilectron.PtrBool(false),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(1000),
					Height:         astilectron.PtrInt(500),
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
					Frame:          astilectron.PtrBool(false),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(1000),
					Height:         astilectron.PtrInt(650),
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
					Frame:          astilectron.PtrBool(false),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(900),
					Height:         astilectron.PtrInt(700),
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
					Frame:          astilectron.PtrBool(false),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(800),
					Height:         astilectron.PtrInt(500),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astilectron.PtrBool(true),
					},
				},
			},
			{
				Homepage:       "ai.html",
				MessageHandler: aiMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astilectron.PtrBool(true),
					Frame:          astilectron.PtrBool(false),
					Modal:          astilectron.PtrBool(false),
					Show:           astilectron.PtrBool(false),
					Fullscreenable: astilectron.PtrBool(false),
					Maximizable:    astilectron.PtrBool(false),
					Minimizable:    astilectron.PtrBool(true),
					Width:          astilectron.PtrInt(800),
					Height:         astilectron.PtrInt(600),
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

func setWindowsShowOrHide(w *astilectron.Window, show bool) {
	if show {
		w.Show()
	} else {
		w.Hide()
	}
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
			timer := time.NewTicker(time.Millisecond * 500)
			for dbPath == "" {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					continue
				}
			}
			timer.Stop()
		} else {
			time.Sleep(time.Second * 2)
		}
		if err := openDB(dbPath); err != nil {
			astilog.Fatal(fmt.Sprintf("running bootstrap failed err=%v", err))
		}
		loadMIBDB()
		go mainWindowBackend(ctx)
		go eventLogger(ctx)
		addEventLog(eventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("TWSNMP起動 データベース='%s'", dbPath),
		})
		go pollingBackend(ctx)
		go logger(ctx)
		go notifyBackend(ctx)
		go arpWatcher(ctx)
		startWindow.Hide()
		mainWindow.Show()
	}()
}
