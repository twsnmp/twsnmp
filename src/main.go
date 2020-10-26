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

	astikit "github.com/asticode/go-astikit"
	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
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
	reportWindow      *astilectron.Window
	aiWindow          *astilectron.Window
	feedbackWindow    *astilectron.Window
	mib               *mibdb.MIBDB
	oui               = &OUIMap{}
	app               *astilectron.Astilectron
	aboutText         = `TWSNMP Manager
Version 5.0.1
Copyright (c) 2019,2020 Masayuki Yamai`
	versionNum = "050001"
	astiLogger *astilog.Logger
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
	logConf := astilog.FlagConfig()
	if logConf.TimestampFormat == "" {
		logConf.TimestampFormat = "01/02 15:04:05.000"
	}
	astiLogger = astilog.New(logConf)
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			astiLogger.Fatalf("could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			astiLogger.Fatalf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			astiLogger.Fatalf("could not create memory profile: %v", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			astiLogger.Fatalf("could not write memory profile:%v", err)
		}
	}
	dbPath = flag.Arg(0)
	if dbPath != "" {
		if err := checkDB(dbPath); err != nil {
			astiLogger.Error(fmt.Sprintf("checkDB(Arg[0]) error=%v", err))
			dbPath = ""
		}
	}
	pctx := context.Background()
	ctx, cancel := context.WithCancel(pctx)

	// Run bootstrap
	astiLogger.Debugf("Running app built at %s", BuiltAt)
	if err := bootstrap.Run(bootstrap.Options{
		Asset:    Asset,
		AssetDir: AssetDir,
		AstilectronOptions: astilectron.Options{
			AppName:            AppName,
			AppIconDarwinPath:  "resources/twsnmp.icns",
			AppIconDefaultPath: "resources/twsnmp.png",
		},
		Debug:  *debug,
		Logger: astiLogger,
		MenuOptions: []*astilectron.MenuItemOptions{{
			Label: astikit.StrPtr("ファイル"),
			SubMenu: []*astilectron.MenuItemOptions{
				{
					Label: astikit.StrPtr("TWSNMPについて"),
					OnClick: func(e astilectron.Event) (deleteListener bool) {
						if err := bootstrap.SendMessage(mainWindow, "about", aboutText, func(m *bootstrap.MessageIn) {
						}); err != nil {
							astiLogger.Error(fmt.Sprintf("sending about event failed err=%v", err))
						}
						return
					},
				},
				{
					Label: astikit.StrPtr("TWSNMPを終了"),
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
			Label: astikit.StrPtr("編集"),
			SubMenu: []*astilectron.MenuItemOptions{
				{
					Label: astikit.StrPtr("全選択"),
					Role:  astilectron.MenuItemRoleSelectAll,
				},
				{
					Label: astikit.StrPtr("切り取り"),
					Role:  astilectron.MenuItemRoleCut,
				},
				{
					Label: astikit.StrPtr("コピー"),
					Role:  astilectron.MenuItemRoleCopy,
				},
				{
					Label: astikit.StrPtr("貼付け"),
					Role:  astilectron.MenuItemRolePaste,
				},
			},
		}, {
			Label: astikit.StrPtr("Window"),
			SubMenu: []*astilectron.MenuItemOptions{
				{
					Checked: astikit.BoolPtr(true), Label: astikit.StrPtr("マップ"),
					Type: astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(mainWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astikit.StrPtr("ノード情報"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(nodeWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astikit.StrPtr("ポーリングリスト"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(pollingListWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astikit.StrPtr("ログ表示"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(logWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astikit.StrPtr("ポーリング"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(pollingWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astikit.StrPtr("MIBブラウザー"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(mibWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astikit.StrPtr("レポート"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(reportWindow, *e.MenuItemOptions.Checked)
						return false
					},
				},
				{
					Label: astikit.StrPtr("AI分析"),
					Type:  astilectron.MenuItemTypeCheckbox,
					OnClick: func(e astilectron.Event) bool {
						setWindowsShowOrHide(aiWindow, *e.MenuItemOptions.Checked)
						if *debug {
							if *e.MenuItemOptions.Checked {
								_ = aiWindow.OpenDevTools()
							} else {
								_ = aiWindow.CloseDevTools()
							}
						}
						return false
					},
				},
			},
		}, {
			Label: astikit.StrPtr("ヘルプ"),
			SubMenu: []*astilectron.MenuItemOptions{{
				Label: astikit.StrPtr("マニュアル"),
				Type:  astilectron.MenuItemTypeNormal,
				OnClick: func(e astilectron.Event) bool {
					_ = openStrURL("https://note.com/twsnmp/m/m15c9aeae6e6d")
					return false
				},
			}, {
				Label: astikit.StrPtr("メールで質問"),
				Type:  astilectron.MenuItemTypeNormal,
				OnClick: func(e astilectron.Event) bool {
					_ = openStrURL("mailto:twsnmp@gmail.com?subject=TWSNMP%20Bug%20Report")
					return false
				},
			}, {
				Label: astikit.StrPtr("フィードバック"),
				Type:  astilectron.MenuItemTypeNormal,
				OnClick: func(e astilectron.Event) bool {
					_ = feedbackWindow.Show()
					return false
				},
			}, {
				Label: astikit.StrPtr("公式ページ"),
				Type:  astilectron.MenuItemTypeNormal,
				OnClick: func(e astilectron.Event) bool {
					_ = openStrURL("https://lhx98.linkclub.jp/twise.co.jp/")
					return false
				},
			}, {
				Label: astikit.StrPtr("最新版ダウンロード"),
				Type:  astilectron.MenuItemTypeNormal,
				OnClick: func(e astilectron.Event) bool {
					_ = openStrURL("https://github.com/twsnmp/twsnmp/releases")
					return false
				},
			},
			}}},
		OnWait: func(a *astilectron.Astilectron, w []*astilectron.Window, m *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
			startWindow = w[0]
			mainWindow = w[1]
			nodeWindow = w[2]
			pollingListWindow = w[3]
			logWindow = w[4]
			pollingWindow = w[5]
			mibWindow = w[6]
			reportWindow = w[7]
			aiWindow = w[8]
			feedbackWindow = w[9]
			app = a
			path := filepath.Join(app.Paths().DataDirectory(), "resources", "mib.txt")
			var err error
			mib, err = mibdb.NewMIBDB(path)
			if err != nil {
				astiLogger.Fatalf("NewMIBDB failed err=%v", err)
			}
			path = filepath.Join(app.Paths().DataDirectory(), "resources", "tlsparams.csv")
			loadTLSParamsMap(path)
			path = filepath.Join(app.Paths().DataDirectory(), "resources", "oui.txt")
			if err := oui.Open(path); err != nil {
				astiLogger.Errorf("OUI Open err=%v", err)
			}
			path = filepath.Join(app.Paths().DataDirectory(), "resources", "services.txt")
			resAppPath = filepath.Join(app.Paths().DataDirectory(), "resources", "app")
			_ = loadServiceMap(path)
			startBackend(ctx)
			mainWindow.On(astilectron.EventNameWindowEventClosed, func(e astilectron.Event) (deleteListener bool) {
				astiLogger.Debug("Main Window Closed")
				app.Stop()
				return
			})
			for i := range w {
				if i < 2 {
					continue
				}
				mi, err := m.Item(2, i-1)
				if err != nil {
					continue
				}
				if w[i] != aiWindow {
					_ = mi.SetVisible(false)
				}
				w[i].On(astilectron.EventNameWindowEventHide, func(e astilectron.Event) (deleteListener bool) {
					_ = mi.SetChecked(false)
					return
				})
				w[i].On(astilectron.EventNameWindowEventMinimize, func(e astilectron.Event) (deleteListener bool) {
					_ = mi.SetChecked(false)
					return
				})
				w[i].On(astilectron.EventNameWindowEventShow, func(e astilectron.Event) (deleteListener bool) {
					_ = mi.SetChecked(true)
					_ = mi.SetVisible(true)
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
					Center:         astikit.BoolPtr(true),
					Modal:          astikit.BoolPtr(true),
					Show:           astikit.BoolPtr(true),
					Frame:          astikit.BoolPtr(false),
					Closable:       astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(false),
					Width:          astikit.IntPtr(450),
					Height:         astikit.IntPtr(500),
				},
			},
			{
				Homepage:       "main.html",
				MessageHandler: mainWindowMessageHandler,
				Options: &astilectron.WindowOptions{
					Center: astikit.BoolPtr(true),
					Show:   astikit.BoolPtr(false),
					Width:  astikit.IntPtr(1024),
					Height: astikit.IntPtr(800),
				},
			},
			{
				Homepage:       "node.html",
				MessageHandler: nodeMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astikit.BoolPtr(true),
					Frame:          astikit.BoolPtr(false),
					Modal:          astikit.BoolPtr(false),
					Show:           astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(true),
					Width:          astikit.IntPtr(850),
					Height:         astikit.IntPtr(450),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astikit.BoolPtr(true),
					},
				},
			},
			{
				Homepage:       "pollingList.html",
				MessageHandler: pollingListMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astikit.BoolPtr(true),
					Frame:          astikit.BoolPtr(false),
					Modal:          astikit.BoolPtr(false),
					Show:           astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(true),
					Width:          astikit.IntPtr(1000),
					Height:         astikit.IntPtr(500),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astikit.BoolPtr(true),
					},
				},
			},
			{
				Homepage:       "log.html",
				MessageHandler: logMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astikit.BoolPtr(true),
					Frame:          astikit.BoolPtr(false),
					Modal:          astikit.BoolPtr(false),
					Show:           astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(true),
					Width:          astikit.IntPtr(1000),
					Height:         astikit.IntPtr(650),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astikit.BoolPtr(true),
					},
				},
			},
			{
				Homepage:       "polling.html",
				MessageHandler: pollingMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astikit.BoolPtr(true),
					Frame:          astikit.BoolPtr(false),
					Modal:          astikit.BoolPtr(false),
					Show:           astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(true),
					Width:          astikit.IntPtr(900),
					Height:         astikit.IntPtr(700),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astikit.BoolPtr(true),
					},
				},
			},
			{
				Homepage:       "mib.html",
				MessageHandler: mibMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astikit.BoolPtr(true),
					Frame:          astikit.BoolPtr(false),
					Modal:          astikit.BoolPtr(false),
					Show:           astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(true),
					Width:          astikit.IntPtr(1200),
					Height:         astikit.IntPtr(800),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astikit.BoolPtr(true),
					},
				},
			},
			{
				Homepage:       "report.html",
				MessageHandler: reportMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astikit.BoolPtr(true),
					Frame:          astikit.BoolPtr(false),
					Modal:          astikit.BoolPtr(false),
					Show:           astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(true),
					Width:          astikit.IntPtr(1200),
					Height:         astikit.IntPtr(980),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astikit.BoolPtr(true),
					},
				},
			},
			{
				Homepage:       "ai.html",
				MessageHandler: aiMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astikit.BoolPtr(true),
					Frame:          astikit.BoolPtr(false),
					Modal:          astikit.BoolPtr(false),
					Show:           astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(true),
					Width:          astikit.IntPtr(800),
					Height:         astikit.IntPtr(600),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astikit.BoolPtr(true),
					},
				},
			},
			{
				Homepage:       "feedback.html",
				MessageHandler: feedbackMessageHandler,
				Options: &astilectron.WindowOptions{
					Center:         astikit.BoolPtr(true),
					Modal:          astikit.BoolPtr(true),
					Show:           astikit.BoolPtr(false),
					Frame:          astikit.BoolPtr(false),
					Closable:       astikit.BoolPtr(false),
					Fullscreenable: astikit.BoolPtr(false),
					Maximizable:    astikit.BoolPtr(false),
					Minimizable:    astikit.BoolPtr(false),
					Width:          astikit.IntPtr(450),
					Height:         astikit.IntPtr(300),
					Custom: &astilectron.WindowCustomOptions{
						HideOnClose: astikit.BoolPtr(true),
					},
				},
			},
		},
	}); err != nil {
		astiLogger.Fatal(fmt.Sprintf("running bootstrap failed err=%v", err))
	}
	stopDiscover()
	cancel()
	closeDB()
	astiLogger.Debug("End of main()")
}

func setWindowsShowOrHide(w *astilectron.Window, show bool) {
	if show {
		_ = w.Show()
	} else {
		_ = w.Hide()
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
				astiLogger.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
				return err.Error(), err
			}
			if err := checkDB(fileName); err != nil {
				astiLogger.Error(fmt.Sprintf("checkDB  error=%v", err))
				return err.Error(), err
			}
			dbPath = fileName
		}
	}
	return "", nil
}

// feedbackMessageHandler handles messages
func feedbackMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "exit":
		_ = feedbackWindow.Hide()
		return "ok", nil
	case "send":
		if len(m.Payload) > 0 {
			var msg string
			if err := json.Unmarshal(m.Payload, &msg); err != nil {
				astiLogger.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
				return err.Error(), err
			}
			sendFeedback(msg)
			_ = feedbackWindow.Hide()
		}
	}
	return "", nil
}

// Backen Process
func startBackend(ctx context.Context) {
	astiLogger.Debug("startBackend")
	go func() {
		if dbPath == "" {
			if err := bootstrap.SendMessage(startWindow, "selectDB", ""); err != nil {
				astiLogger.Error(fmt.Sprintf("sendSendMessage selectDB error=%v", err))
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
			astiLogger.Fatal(fmt.Sprintf("running bootstrap failed err=%v", err))
		}
		_ = setupInfluxdb()
		_ = setupRestAPI()
		_ = loadMIBDB()
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
		go aiWindowBackend(ctx)
		go reportBackend(ctx)
		_ = startWindow.Hide()
		_ = mainWindow.Show()
		_ = mainWindow.Resize(mainWindowInfo.Width, mainWindowInfo.Height)
		if mainWindowInfo.Top < 0 {
			_ = mainWindow.Center()
		} else {
			_ = mainWindow.Move(mainWindowInfo.Left, mainWindowInfo.Top)
		}
	}()
}
