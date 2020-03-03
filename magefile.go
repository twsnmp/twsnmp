// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var pkgList = []string{
	"github.com/jessevdk/go-flags",
	"github.com/asticode/go-astilectron",
	"github.com/asticode/go-astilectron-bundler/...",
	"github.com/asticode/go-astilectron-bootstrap/...",
	"github.com/asticode/go-astikit",
	"github.com/twsnmp/go-mibdb",
	"go.etcd.io/bbolt",
	"github.com/signalsciences/ipv4",
	"gopkg.in/mcuadros/go-syslog.v2",
	"github.com/tehmaze/netflow",
	"github.com/soniah/gosnmp",
	"github.com/twsnmp/go-mibdb",
	"github.com/konsorten/go-windows-terminal-sequences",
	"github.com/mattn/go-colorable",
	"github.com/robertkrimen/otto",
	"github.com/vjeantet/grok",
	"github.com/mrichman/godnsbl",
	"github.com/openrdap/rdap",
	"github.com/asticode/go-astilog",
}

// Build 実行ファイルのビルド
func Build() error {
	fmt.Println("Building... ")
	return buildInternal(true)
}

// BuildMac Mac用の実行ファイルのビルド
func BuildMac() error {
	fmt.Println("Building... ")
	return buildInternal(false)
}

func buildInternal(bWindows bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cwd)
	err = os.Chdir("./src")
	if err != nil {
		return err
	}
	if bWindows {
		err = sh.RunV("astilectron-bundler", "-w")
	} else {
		err = sh.RunV("astilectron-bundler")
	}
	if err != nil {
		return err
	}
	return nil
}

// InstallDeps  ビルドに必要なパッケージのインストール
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	for _, p := range pkgList {
		err := sh.Run("go", "get", p)
		if err != nil {
			return err
		}
	}
	return nil
}

// MakeZip リリース用のZIPファイルを作成
func MakeZip() error {
	mg.Deps(Build)
	fmt.Println("Make ZIP...")
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cwd)
	if _, err := os.Stat("./rel"); os.IsNotExist(err) {
		os.Mkdir("./rel", 0777)
	}
	err = os.Chdir("./src")
	if err != nil {
		return err
	}
	err = os.Chdir("./output/windows-amd64")
	if err != nil {
		return err
	}
	err = sh.RunV("zip", "-r", "../../../rel/TwsnmpWin.zip", ".")
	if err != nil {
		return err
	}
	err = os.Chdir("../darwin-amd64")
	if err != nil {
		return err
	}
	err = sh.RunV("zip", "-r", "../../../rel/TwsnmpMacOS.zip", ".")
	if err != nil {
		return err
	}
	return nil
}

// UpdateDeps  ビルドに必要なパッケージのアップデート
func UpdateDeps() error {
	fmt.Println("Installing Deps...")
	for _, p := range pkgList {
		err := sh.Run("go", "get", "-u", p)
		if err != nil {
			return err
		}
	}
	return nil
}

// Clean ビルドした実行ファイルの削除
func Clean() {
	fmt.Println("Cleaning...")
	sh.Run("/bin/sh", "-c", "rm -rf  ./rel/*  ./src/output/*")
}
