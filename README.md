Tavle
=======================================

[![Build Status](https://travis-ci.org/ryumei/tavle.svg?branch=master)](https://travis-ci.org/ryumei/tavle)

A tiny chat server for classroom.

## 概要 Description

It is an example implementation of online chat system using WebSocket.
*tavle* is a blackboard in Norwegian.

Websocket を使ったチャットの実装例です。
このツールは、コミュニティ [eLV](http://www.elv.tokyo/) の活動の一環で生まれました。

とりあえず動かしてみたい方は [How to Use](HOW_TO_USE.md) をご覧ください。

## 必要なもの Requirement

サーバの実行は、Windows、Mac、Linux 上で動きます。
クライアントはウェブブラウザ (Chrome、Safari、Firefoxなど) で接続します。

## 使い方 Usage

設定ファイルを配置し、実行します。
[How to Use](HOW_TO_USE.md) をご覧ください。

## インストール方法 Install

ビルド済みのバイナリを入手し、展開してください。
ビルドする場合は、次の節も参考にしてください。

## 開発者向け情報 How to Build

開発作業確認済み環境情報

* Go (v1.9)
* glide (v0.12.3)

### ビルド方法

```
$ make
```

クロスコンパイル

```
$ make dist
```

## Contribution

1. Fork
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Run test suite with the `go test ./...` command and confirm that it passes
6. Run `gofmt -s`
7. Create new Pull Request

## ライセンス License

[Apache License Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

## Author

Takaaki NAKAJIMA
