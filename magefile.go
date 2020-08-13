// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Build : 実行ファイルのビルド
func Build() error {
	fmt.Println("Building... ")
	return buildInternal(true)
}

// BuildMac : Mac用の実行ファイルのビルド
func BuildMac() error {
	fmt.Println("Building... ")
	return buildInternal(false)
}

func buildInternal(bAll bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cwd)
	err = os.Chdir("./src")
	if err != nil {
		return err
	}
	if bAll {
		err = sh.RunV("astilectron-bundler")
	} else {
		err = sh.RunV("astilectron-bundler", "-c", "bundlerMac.json")
	}
	if err != nil {
		return err
	}
	return nil
}

// MakeZip : リリース用のZIPファイルを作成
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
	for s, d := range map[string]string{
		"windows-amd64": "TwsnmpWin.zip",
		"darwin-amd64":  "TwsnmpMacOS.zip",
		"linux-amd64":   "TwsnmpLinuxAMD64.zip",
		"linux-arm":     "TwsnmpLinuxArm.zip",
	} {
		if err := makeOneZip(s, d); err != nil {
			return err
		}
	}
	return nil
}

func makeOneZip(src, dst string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cwd)
	err = os.Chdir("./src/output/" + src)
	if err != nil {
		return err
	}
	err = sh.RunV("zip", "-r", "../../../rel/"+dst, ".")
	if err != nil {
		return err
	}
	return nil
}

// Clean : ビルドした実行ファイルの削除
func Clean() {
	fmt.Println("Cleaning...")
	sh.Run("/bin/sh", "-c", "rm -rf  ./rel/*  ./src/output/*")
}
