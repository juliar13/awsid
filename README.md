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

#### アカウント情報の自動取得

このツールは AWS Organizations API を使用してアカウント情報を自動的に取得し、`~/.aws/account_info` ファイルにCSV形式で保存します。

**重要**: AWS Organizations API の呼び出しには us-east-1 リージョンが使用されます。これは AWS Organizations がグローバルサービスであり、標準的なリージョンとして us-east-1 が推奨されているためです。

#### 手動設定（オプション）

必要に応じて、`~/.aws/account_info` ファイルを手動で編集することも可能です：

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

アカウント名で検索（--nameオプション）：

```bash
awsid --name test
# test を含むアカウント名で検索
```

結果をソート：

```bash
awsid --sort name              # アカウント名昇順
awsid --sort-desc name         # アカウント名降順
awsid --sort joined_timestamp  # 作成日昇順
awsid --sort-desc joined_timestamp  # 作成日降順（新しい順）
```

## 出力形式

出力形式は以下の方法で指定できます：

### 統一フォーマットオプション（推奨）

```bash
awsid --format json    # JSON形式
awsid --format table   # テーブル形式  
awsid --format csv     # CSV形式
```

### 個別フォーマットフラグ（下位互換性）

```bash
awsid --json          # JSON形式
awsid --table         # テーブル形式
awsid --csv           # CSV形式
```

**注意**: `--format`オプションと個別フラグが同時に指定された場合、`--format`が優先されます。

## ソート機能

結果は以下のフィールドでソートできます：

### ソートオプション

```bash
awsid --sort <field>        # 昇順ソート
awsid --sort-desc <field>   # 降順ソート
```

### 利用可能なソートフィールド

- `id` - アカウントID
- `name` - アカウント名
- `email` - メールアドレス
- `status` - ステータス
- `joined_timestamp` - 作成日時
- `joined_method` - 参加方法

### ソート例

```bash
# アカウント名でアルファベット順
awsid --sort name --format table

# 最新作成順
awsid --sort-desc joined_timestamp

# 検索結果をソート
awsid --name test --sort name

# JSON形式でメールアドレス順
awsid --sort email --format json
```

**注意**: `--sort`と`--sort-desc`は同時に指定できません。

### 標準出力（デフォルト）

完全一致の場合はアカウントIDのみ：
```bash
awsid yamasaki-test
# 出力: 123456789012
```

部分一致や全表示の場合は詳細情報：
```bash
awsid yamasaki
# 出力:
# ID: 123456789012 | ARN: arn:aws:organizations::... | Email: ... | Name: yamasaki-test | Status: ACTIVE | Method: CREATED | Joined: 2024-01-01T...
```

### JSON形式

```bash
awsid yamasaki --format json
# または
awsid yamasaki --json
# 出力:
# {
#     "account_info": [
#         {
#             "id": "123456789012",
#             "arn": "arn:aws:organizations::...",
#             "email": "test@example.com",
#             "name": "yamasaki-test",
#             "status": "ACTIVE",
#             "joined_method": "CREATED",
#             "joined_timestamp": "2024-01-01T...",
#             "alias_name": "yamasaki-test",
#             "account_id": "123456789012"
#         }
#     ]
# }
```

### テーブル形式

```bash
awsid yamasaki --format table
# または  
awsid yamasaki --table
# 出力:
# ┌──────────────┬─────────────────┬───────────────────┬───────────────┬────────┬───────────────┬──────────────────┐
# │      ID      │       ARN       │       EMAIL       │     NAME      │ STATUS │ JOINED METHOD │ JOINED TIMESTAMP │
# ├──────────────┼─────────────────┼───────────────────┼───────────────┼────────┼───────────────┼──────────────────┤
# │ 123456789012 │ arn:aws:org...  │ test@example.com  │ yamasaki-test │ ACTIVE │ CREATED       │ 2024-01-01T...   │
# └──────────────┴─────────────────┴───────────────────┴───────────────┴────────┴───────────────┴──────────────────┘
```

### CSV形式

```bash
awsid yamasaki --format csv
# または
awsid yamasaki --csv
# 出力:
# id,arn,email,name,status,joined_method,joined_timestamp
# 123456789012,arn:aws:organizations::...,test@example.com,yamasaki-test,ACTIVE,CREATED,2024-01-01T...
```

## ライセンス

MIT
