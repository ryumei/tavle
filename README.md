Tavle
=======================================

A tiny chat server for classroom.

## 概要 Description

It is an example implementation of online chat system using WebSocket.
*tavle* is a blackboard in Norwegian.

Websocket を使ったチャットの実装例です。

## 必要なもの Requirement

サーバの実行は、Windows、Mac、Linux 上で動きます。
クライアントはウェブブラウザ (Chrome、Safari、Firefoxなど) で接続します。

## 使い方 Usage

設定ファイルを配置し、実行します。
サーバ役ホストの IP アドレスを調べておいてください。

### 初期設定

``tavle.tml.sample`` を ``tavle.tml`` という名前にコピーし、
[Server] の Endpoint と Port を適切に修正してください (設定項目は大文字小文字を区別します)。

### サーバの起動

    $ ./tavle -c tavle.tml

Windows の場合は、 実行ファイルは tavle.exe という名称です。
tavle.tml が同じディレクトリにある場合には、``-c`` オプションは省略可能です。

### クライアントからの接続

起動の後、ブラウザで http://SERVER_IP:8888/ にアクセスしてみてください。
ポート番号は、設定ファイルにて指定したものに読み替えてください。

ログイン画面が開きますので、入力してください。

* ユーザ名: 必須です。
* メールアドレス: メールアドレスは任意です (gravater と連携します)。
* ルーム名: 省略すると、デフォルトのホワイエ (foyer) に入ります。


#### [TroubleShooting] 接続できない時

サーバおよび、クライアントのファイアウォールを確認してください。

  * セキュリティソフトウェア
  * OS
  * ネットワーク経路

#### [KnownIssue] ブラウザの再読み込みで、チャット履歴が消える。

ブラウザ再読み込みすると、画面がクリアされます。
(送信済みのメッセージは、相手側では見えたままです)


### SSL/TLS 証明書

設定ファイルの EnableTLS を true
OpenSSL で作るサンプルを contrib/create_certs.sh に。

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
