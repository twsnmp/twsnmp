# twsnmp
TWSNMPマネージャ 復刻版
(TWSNMP Manager Returns)

2020年以降に開発は中止しました。
2023年にリポジトリをアーカイブしました。

[![Godoc Reference](https://godoc.org/github.com/twsnmp/twsnmp/src?status.svg)](http://godoc.org/github.com/twsnmp/twsnmp/src)
[![Build Status](https://travis-ci.org/twsnmp/twsnmp.svg?branch=master)](https://travis-ci.org/twsnmp/twsnmp)
[![Go Report Card](https://goreportcard.com/badge/twsnmp/twsnmp)](https://goreportcard.com/report/twsnmp/twsnmp)

## Overview

1999年に開発し、今でも多くのユーザーが利用しているTWSNMPマネージャを2019年の技術で復活させるプロジェクトです。

This project is to revive the TWSNMP Manager that was developed in 1999 and is still used by many users with the technology of 2019.

![TWSNMP](http://www.twise.co.jp/img/twsnmp_title.jpg)

## Status


以下の機能を実装しましたが、開発は中止しました。

* マップ表示、編集
* 自動発見
* ノードの追加、編集
* ポーリングの追加、編集
* ノード情報の表示
* PING監視
* SNNP監視
* Syslog受信
* Netflow受信
* Trap受信
* MIBブラウザー
* ログ検索表示
* メール通知
* ログ監視
* ポーリング結果のグラフ表示
* DNS監視
* HTTP/HTTPS監視
* TCP監視
* TLS監視
* AIによる異常診断
* デバイス、ユーザー、サーバー、通信フローの信用スコアレポート
* VMware仮想基盤の監視
* TWSNMP間の連携、Webブラウザーでマップ表示
* Influxdb経由でGrafana連携
* ローカル、SSH経由のコマンド実行監視
* 定期レポート、復帰通知メール(v5.0.1)
* 更新版の確認、フィードバック機能(v5.0.1)

以下の動作環境に対応しました。

* Windows
* Mac OS
* Linux AMD64
* Linux ARM (Raspberry Pi 4)

![TWSNMP](https://assets.st-note.com/production/uploads/images/15118776/picture_pc_9d8f9c01141ab53f0c72ec1384cb36c4.png)

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
  makeZip        リリース用のZIPファイルを作成
```

```
$mage build
```
を実行すれば、MacOS,Windows,Linux(amd64),Linux(arm)用の実行ファイルが、`src/output`のディレクトリに作成されます。

配布用のZIPファイルを作成するためには、
```
$mage makeZip
```
を実行します。ZIPファイルが`rel/`ディレクトリに作成されます。

## Block図

内部構造は、以下の図の用になっています。概ねソースコードのファイル名と機能ブロックは対応しています。

![TWSNMP](https://d2l930y2yx77uc.cloudfront.net/production/uploads/images/20504835/picture_pc_2f2b09a18c74cfd6f7a2aebfeb9dc096.png)

## Copyright

see ./LICENSE

```
Copyright 2019,2020 Masayuki Yamai
```
