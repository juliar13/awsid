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

`~/.aws/account_info` ファイルに以下の形式でアカウント情報を記載します：

```
# AWS Account information
# Format: alias_name account_id
yamasaki-test 123456789012
yamasaki-test-dev 123456789013
yamasaki-prod 123456789014
other-account 987654321098
```

### コマンド

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

JSON形式で出力：

```bash
awsid yamasaki-test --json
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
#         }
#     ]
# }
```

## ライセンス

MIT
