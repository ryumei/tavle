A talk board

*tavle* is a blackboard in Norwegian.

## 初期設定

tavle.tml.sample を tavle.tml という名前にコピーし、
[Server] の Port を適切に修正してください (設定項目は大文字小文字を区別します)。

```tml
[Server]
Port   = 8888
Endpoint = ""
Debug = false

[Log]
AccessLog = "access.log"
ServerLog = "server.log"
Level = "INFO"
```

## 起動の仕方

    $ ./tavle -c tavle.tml

tavle.tml が同じディレクトリにある場合には、``-c`` オプションは省略可能です。

## 接続の仕方

起動の後、ブラウザで http://localhost:8888/ にアクセスしてみてください。
ポート番号は、設定ファイルにて指定したものに読み替えてください。

### [TroubleShooting] 接続できない時

ファイアウォールを確認してください

  * セキュリティソフトウェア
  * OS
  * ネットワーク経路

