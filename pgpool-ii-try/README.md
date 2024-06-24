# pgpool-II try

pgpool2 を試す。

OS: Ubuntu22.04
Postgresql 14
pgpool2 4.1

## 事前準備

### vagrant

virtualbox を使う

#### 起動

```bash
vagrant up
```

#### 終了

```bash
vagrant halt
```

#### 削除

```bash
vagrant destroy
```

### ansible

ansible は pipenv でインストールする

#### 初回

```bash
pipenv install
```

#### 毎回

```bash
pipenv shell
```

## ログ

### とりあえず postgresql-14 と pgpool2 をインストール

```bash
ansible-playbook -i hosts install-playbook.yml
```

特にリポジトリを追加せずにデフォルトのままのため、バージョンは postgres が 14、pgpool2 が 4.1.4 がインストールされた。

```bash
vagrant@server1:~$ pgpool -v
pgpool-II version 4.1.4 (karasukiboshi)
vagrant@server1:~$ psql --version
psql (PostgreSQL) 14.12 (Ubuntu 14.12-0ubuntu0.22.04.1)
```
