コマンド実行時に AWS Organizations にアクセスし、
`aws organizations list-accounts` で取得できる各アカウントの情報をCSV形式で`~/.aws/account_info`に保存したい。
現在は alias_name,account_id だけなので、それを拡張する形になります。
以下は取得したJSONの例です。
```json
        {
            "Id": "058264188814",
            "Arn": "arn:aws:organizations::691610352964:account/o-8pr5zme5bo/058264188814",
            "Email": "judydoctrine+iam-bation@gmail.com",
            "Name": "pjuliar-accountname-iam-bastion",
            "Status": "ACTIVE",
            "JoinedMethod": "CREATED",
            "JoinedTimestamp": "2024-02-24T13:08:50.690000+09:00"
        }
```

上記のJSONの例をCSVにすると以下のようになります。

```csv
id,arn,email,name,status,joined_method,joined_timestamp
058264188814,arn:aws:organizations::691610352964:account/o-8pr5zme5bo/058264188814,judydoctrine+iam-bation@gmail.com,pjuliar-accountname-iam-bastion,ACTIVE,CREATED,2024-02-24T13:08:50.690000+09:00
```
