# AWSID

AWSアカウントのエイリアス名からアカウントIDを出力するためのCLIツール

## インストール

Homebrewを使用してインストールできます：

```bash
brew tap juliar13/awsid
brew install awsid
```

## 使い方

### 設定

`~/.aws/account_info` ファイルにCSV形式でアカウント情報を記載します：

```csv
alias_name,account_id
yamasaki-test,123456789012
yamasaki-test-dev,123456789013
yamasaki-prod,123456789014
other-account,987654321098
```

CSV形式はExcelなどのスプレッドシートアプリケーションからのインポート・エクスポートが容易で、データ管理が効率的です。

### コマンド

バージョンを確認：

```bash
awsid --version
# 出力: awsid version 0.2.0
```

全てのアカウント情報を表示：

```bash
awsid
```

特定のエイリアス名に対応するアカウントIDを表示：

```bash
awsid yamasaki-test
# 出力: 123456789012
```

特定の文字列で始まるエイリアス名を持つアカウント情報を表示：

```bash
awsid yamasaki
# 出力:
# yamasaki-test: 123456789012
# yamasaki-test-dev: 123456789013
# yamasaki-prod: 123456789014
```

## 出力形式

### 標準出力（デフォルト）

完全一致の場合はアカウントIDのみ：
```bash
awsid yamasaki-test
# 出力: 123456789012
```

部分一致や全表示の場合は「エイリアス名: アカウントID」形式：
```bash
awsid yamasaki
# 出力:
# yamasaki-test: 123456789012
# yamasaki-test-dev: 123456789013
# yamasaki-prod: 123456789014
```

### JSON形式（--jsonフラグ）

```bash
awsid yamasaki --json
# 出力:
# {
#     "account_info": [
#         {
#             "alias_name": "yamasaki-test",
#             "account_id": "123456789012"
#         },
#         {
#             "alias_name": "yamasaki-test-dev", 
#             "account_id": "123456789013"
#         },
#         {
#             "alias_name": "yamasaki-prod",
#             "account_id": "123456789014"
#         }
#     ]
# }
```

### テーブル形式（--tableフラグ）

```bash
awsid yamasaki --table
# 出力:
# +-------------------+----------------+
# | ALIAS NAME        | ACCOUNT ID     |
# +-------------------+----------------+
# | yamasaki-test     | 123456789012   |
# | yamasaki-test-dev | 123456789013   |
# | yamasaki-prod     | 123456789014   |
# +-------------------+----------------+
```

### CSV形式（--csvフラグ）

```bash
awsid yamasaki --csv
# 出力:
# alias_name,account_id
# yamasaki-test,123456789012
# yamasaki-test-dev,123456789013
# yamasaki-prod,123456789014
```

## ライセンス

MIT
