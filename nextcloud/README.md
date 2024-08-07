# nextcloud試す

## 起動

```bash
docker compose up -d
```

## セットアップ

http://localhost:8080/ にアクセスして管理ユーザーの作成を行う。

## メモ

設定項目のメモ

### 共有設定

- アプリからの共有APIの利用を許可する
    - 他ユーザー/グループへやURLでの共有ができるようになる
    - 再共有を許可する: 共有されたユーザーが更に共有できるようになる
    - グループ共有を許可する: グループに対して共有できる
    - グループ内のユーザーでのみ共有するように制限する: ?

- URLリンクとメールでの共有を許可する
    - URLで共有できるようになる

- デフォルトの共有アクセス許可
    - 選択した許可内容がデフォルトで選択される。共有時に変更可能。

### セキュリティ

- 二要素認証
    - Two-Factor TOTP Providerを有効化することで利用可能

- サーバーサイド暗号化
    - アップロードされたファイルを暗号化して保存する
    - default encryption moduleというアプリを有効化する必要がある(インストールはされている)
    - 有効後にアップロードされるファイルが対象となり、既存のものは暗号化されない(上書き保存すると暗号化される)

### 外部ストレージ

- External storage supportを有効化することで利用可能
- s3やftpなど外部ストレージに接続できる
- 接続したストレージはフォルダとしてユーザーに見える
- 利用できるユーザー・グループを指定できる

## バックアップ/リストア

### バックアップ

https://docs.nextcloud.com/server/latest/admin_manual/maintenance/backup.html

メンテナンスモードにしておく。
終了時にoffにする。

`-u 33`はwww-dataユーザーで実行

```bash
docker compose exec -it -u 33 app bash
php occ maintenance:mode --on
```

データのバックアップは `config/` `data/` `themes/`を対象にしていれば良さそう。またはnextcloudのディレクトリ全体。

```bash
rsync -Aavx nextcloud/ nextcloud-dirbkp_`date +"%Y%m%d"`/
```


DBのバックアップはmysqldumpで出力する。

```bash
mysqldump --single-transaction -h [server] -u [username] -p[password] [db_name] > nextcloud-sqlbkp_`date +"%Y%m%d"`.bak
```