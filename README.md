A talk board

*tavle* is a blackboard in Norwegian.

# 使い方

設定ファイルを配置し、実行します。

## 初期設定

tavle.tml.sample を tavle.tml という名前にコピーし、
[Server] の Port を適切に修正してください (設定項目は大文字小文字を区別します)。

```tml
[Server]
Port   = 8888
Endpoint = ""
Debug = false
EnableTLS = false
CertFile = ""
KeyFile = ""

[Log]
AccessLog = "access.log"
ServerLog = "server.log"
Level = "INFO"
```

## 起動の仕方

    $ ./tavle -c tavle.tml

Windows の場合は、 実行ファイルは tavle.exe という名称です。
tavle.tml が同じディレクトリにある場合には、``-c`` オプションは省略可能です。

## 接続の仕方

起動の後、ブラウザで http://localhost:8888/ にアクセスしてみてください。
ポート番号は、設定ファイルにて指定したものに読み替えてください。

### [TroubleShooting] 接続できない時

ファイアウォールを確認してください

  * セキュリティソフトウェア
  * OS
  * ネットワーク経路

# SSL/TLS 証明書

OpenSSL で作る例。

## 認証局役

```
openssl genrsa -out ca-privatekey.pem 2048
openssl req -new -key ca-privatekey.pem -out ca-csr.pem
openssl req -x509 -key ca-privatekey.pem -in ca-csr.pem -out ca-crt.pem -days 3650
```

## サーバ役

```
openssl genrsa -out server-privatekey.pem
openssl req -new -key server-privatekey.pem -out server-csr.pem
openssl x509 -req -CA ca-crt.pem -CAkey ca-privatekey.pem -CAcreateserial -in server-csr.pem -out server-crt.pem -days 3650
```
