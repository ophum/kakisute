resource "sakuracloud_disk" "bench_disk" {
  name              = "bench-disk"
  source_archive_id = data.sakuracloud_archive.ubuntu.id
  size              = 40
}

resource "sakuracloud_server" "bench_server" {
  name = "bench"
  tags = ["__with_sacloud_inventory"]
  description = jsonencode({
    sacloud_inventory = {
      hostname_type = "nic0_ip",
      host_vers = {
        nic1_ip = cidrhost("10.0.0.0/24", 4)
      }
    },
  })
  disks = [sakuracloud_disk.bench_disk.id]

  network_interface {
    upstream = "shared"
  }
  network_interface {
    upstream = sakuracloud_switch.sw.id
  }

  user_data = templatefile("user-data.yaml", {
    fqdn = "bench"
  })
}

