# twsnmp
TWSNMPマネージャ 復刻版
(TWSNMP Manager Returns)

[![Godoc Reference](https://godoc.org/github.com/twsnmp/twsnmp/src?status.svg)](http://godoc.org/github.com/twsnmp/twsnmp/src)
[![Build Status](https://travis-ci.org/twsnmp/twsnmp.svg?branch=master)](https://travis-ci.org/twsnmp/twsnmp)
[![Go Report Card](https://goreportcard.com/badge/twsnmp/twsnmp)](https://goreportcard.com/report/twsnmp/twsnmp)

## Overviewp

1999年に開発し、今でも多くのユーザーが利用しているTWSNMPマネージャを2019年の技術で復活させるプロジェクトです。

This project is to revive the TWSNMP Manager that was developed in 1999 and is still used by many users with the technology of 2019.

![TWSNMP](http://www.twise.co.jp/img/twsnmp_title.jpg)

## Status

アプリケーションのビルドができて、画面の操作イメージがわかるレベルです。

![TWSNMP](https://d2l930y2yx77uc.cloudfront.net/production/uploads/images/14555554/picture_pc_b92f0d0955e7d448a1f2151d6e16920f.png)

## Build


ビルドには、Mageを利用します。
https://magefile.org

次のコマンドで、インストールします。

```
go get -u -d github.com/magefile/mage
cd $GOPATH/src/github.com/magefile/mage
go run bootstrap.go
```

以下のターゲットがmageコマンドで指定できます。

```
$ mage
Targets:
  build          実行ファイルのビルド
  buildMac       Mac用の実行ファイルのビルド
  clean          ビルドした実行ファイルの削除
  installDeps    ビルドに必要なパッケージのインストール
  makeZip        リリース用のZIPファイルを作成
  updateDeps     ビルドに必要なパッケージのアップデート
```

初回のビルド前に、
```
$mage installDeps
```
を実行して、ビルドに必要なパッケージをインストールします。

その後、
```
$mage build
```
を実行すれば、MacOS,Windows用の実行ファイルが、`src/output`のディレクトリに作成されます。

配布用のZIPファイルを作成するためには、
```
$mage makeZip
```
を実行します。ZIPファイルが`rel/`ディレクトリに作成されます。

## Copyright

see ./LICENSE

```
Copyright 2019 Masayuki Yamai
```
