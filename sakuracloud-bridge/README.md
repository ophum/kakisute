# sakuracloud bridge

##### 2024/08/12 (月)

さくらのクラウドのブリッジを試す。

is1bとtk1bにスイッチを作成しブリッジで接続する

各ゾーンにサーバーを作成し、サーバー間でpingなどを行いテストする。

### 準備

```bash
git clone git@github.com:ophum/kakisute.git
cd kakisute/sakuracloud-bridge/

terraform init
terraform plan
terraform apply
```

適宜クラウドのコンパネやusacloudでグローバルIPアドレスを確認しsshする

```bash
usacloud iaas server list --zone all
```

### ens4にipアドレスを設定する

ブリッジで繋いでいるスイッチは10.0.0.0/24のネットワークとします。
またこのネットワークに接続しているインターフェースはens4となります。

is1b

```bash
sudo ip address add 10.0.0.1/24 dev ens4
sudo ip link set up dev ens4
```

tk1b

```bash
sudp ip address add 10.0.0.2/24 dev ens4
sudo ip link set up dev ens4
```

### tk1bからis1bにpingを実行する

```bash
ubuntu@tk1b:~$ ping 10.0.0.1 -c 4
PING 10.0.0.1 (10.0.0.1) 56(84) bytes of data.
64 bytes from 10.0.0.1: icmp_seq=1 ttl=64 time=37.1 ms
64 bytes from 10.0.0.1: icmp_seq=2 ttl=64 time=18.4 ms
64 bytes from 10.0.0.1: icmp_seq=3 ttl=64 time=18.4 ms
64 bytes from 10.0.0.1: icmp_seq=4 ttl=64 time=18.4 ms

--- 10.0.0.1 ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 3005ms
rtt min/avg/max/mdev = 18.385/23.067/37.075/8.087 ms
```

### is1bに1GBのファイルを設置httpで公開し、tk1bでcurlで取得する

```bash
ubuntu@is1b:~$ dd if=/dev/random of=./data count=1024 bs=1M
1024+0 records in
1024+0 records out
1073741824 bytes (1.1 GB, 1.0 GiB) copied, 6.49598 s, 165 MB/s
ubuntu@is1b:~$ ls -lah
total 1.1G
drwxr-x--- 4 ubuntu ubuntu 4.0K Aug 12 04:31 .
drwxr-xr-x 3 root   root   4.0K Aug 12 04:27 ..
-rw-r--r-- 1 ubuntu ubuntu  220 Jan  6  2022 .bash_logout
-rw-r--r-- 1 ubuntu ubuntu 3.7K Jan  6  2022 .bashrc
drwx------ 2 ubuntu ubuntu 4.0K Aug 12 04:29 .cache
-rw-r--r-- 1 ubuntu ubuntu  807 Jan  6  2022 .profile
drwx------ 2 ubuntu ubuntu 4.0K Aug 12 04:27 .ssh
-rw-r--r-- 1 ubuntu ubuntu    0 Aug 12 04:30 .sudo_as_admin_successful
-rw-rw-r-- 1 ubuntu ubuntu 1.0G Aug 12 04:31 data
ubuntu@is1b:~$ python3 -m http.server
Serving HTTP on 0.0.0.0 port 8000 (http://0.0.0.0:8000/) ...
```

```bash
ubuntu@tk1b:~$ curl -w "speed: %{speed_download}\n" -o /dev/null -s http://10.0.0.1:8000/data
speed: 111949187
```

111949187 B/s

だいたい111MB/s

1Gbps出てる感じ

さくらのクラウドのスイッチの帯域は接続するサーバーのスペックに依存します。
今回作成したサーバーはメモリ1GBで32GB未満なので1Gbpsが制限値となります。
また、ブリッジ接続には帯域制限がないのでしっかりサーバーに応じた制限がかかっていることが分かりました。

[スイッチに帯域制限はありますか?](https://manual.sakura.ad.jp/cloud/support/technical/network.html#support-network-03)


### サーバーのメモリを32GBにして帯域を2Gbpsにしてみる

32GBにすれば帯域は2Gbpsになるので試してみます。

terraformで5コアメモリを32GBにします。(5コア以上のプランのみ32GBを選べる)
```diff
$ diff -u main.tf.old main.tf
--- main.tf.old 2024-08-12 13:55:24.574689621 +0900
+++ main.tf     2024-08-12 13:58:34.214697924 +0900
@@ -38,6 +38,8 @@
 resource "sakuracloud_server" "is1b_server" {
   name        = "is1b_server"
   disks       = [sakuracloud_disk.is1b_disk.id]
+  core = 5
+  memory = 32
 
   network_interface {
     upstream         = "shared"
@@ -56,6 +58,8 @@
 resource "sakuracloud_server" "tk1b_server" {
   name        = "tk1b_server"
   disks       = [sakuracloud_disk.tk1b_disk.id]
+  core = 5
+  memory = 32
 
   network_interface {
     upstream         = "shared"
```

terraform applyするとサーバーが再作成されます。

sshしもう一度ens4にipアドレスを設定しておきます。

is1bでhttp.serverを起動しtk1bでcurlします。

```bash
ubuntu@tk1b:~$ curl -w "speed: %{speed_download}\n" -o /dev/null -s http://10.0.0.1:8000/data
speed: 166830998
```

166830998 B/s

だいたい166 MB/s


2Gbpsとまではいかないけど1Gbpsよりは速い。

### 同一ゾーン内でcurlしてみる

以下のようにis1bにもう1台5コア32GBメモリのサーバーを追加し同一ゾーン内でcurlしてみる。
```
$ diff -u main.tf.old main.tf
--- main.tf.old 2024-08-12 14:11:31.224687173 +0900
+++ main.tf     2024-08-12 14:11:50.594686653 +0900
@@ -55,6 +55,27 @@
   zone = "is1b"
 }
 
+resource "sakuracloud_server" "is1b_server2" {
+  name        = "is1b_server2"
+  disks       = [sakuracloud_disk.is1b_disk2.id]
+  core = 5
+  memory = 32
+
+  network_interface {
+    upstream         = "shared"
+  }
+
+  network_interface {
+    upstream = sakuracloud_switch.is1b.id
+  }
+
+  user_data = templatefile("user-data.yaml", {
+    fqdn = "is1b2"
+  })
+  zone = "is1b"
+}
+
+
 resource "sakuracloud_server" "tk1b_server" {
   name        = "tk1b_server"
   disks       = [sakuracloud_disk.tk1b_disk.id]
@@ -103,6 +124,12 @@
   source_archive_id = data.sakuracloud_archive.is1b_ubuntu.id
   zone = "is1b"
 }
+
+resource "sakuracloud_disk" "is1b_disk2" {
+  name              = "is1b_disk2"
+  source_archive_id = data.sakuracloud_archive.is1b_ubuntu.id
+  zone = "is1b"
+}
 
 resource "sakuracloud_disk" "tk1b_disk" {
   name              = "tk1b_disk"
```

追加したサーバーにsshしてipを設定しpingとcurlしてみます。

ping

```bash
ubuntu@is1b2:~$ ping 10.0.0.1 -c 4
PING 10.0.0.1 (10.0.0.1) 56(84) bytes of data.
64 bytes from 10.0.0.1: icmp_seq=1 ttl=64 time=0.911 ms
64 bytes from 10.0.0.1: icmp_seq=2 ttl=64 time=0.241 ms
64 bytes from 10.0.0.1: icmp_seq=3 ttl=64 time=0.214 ms
64 bytes from 10.0.0.1: icmp_seq=4 ttl=64 time=0.218 ms

--- 10.0.0.1 ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 3066ms
rtt min/avg/max/mdev = 0.214/0.396/0.911/0.297 ms
```

curl 
```bash
ubuntu@is1b2:~$ curl -w "speed: %{speed_download}\n" -o /dev/null -s http://10.0.0.1:8000/data
speed: 258703608
```
258703608 B/s

大体258MB/s = 2000Mb/s

同一ゾーンであればしっかり2GBps出ました。
