data "sakuracloud_archive" "ubuntu" {
  filter {
    tags = [
      "cloud-init",
      "distro-ubuntu",
      "distro-ver-22.04.5",
    ]
  }
}

resource "sakuracloud_disk" "disk" {
  count             = 3
  name              = format("disk%d", count.index)
  source_archive_id = data.sakuracloud_archive.ubuntu.id
  size              = 40
}

resource "sakuracloud_server" "server" {
  count = 3
  name  = format("etcd%d", count.index + 1)
  tags  = ["__with_sacloud_inventory"]
  description = jsonencode({
    sacloud_inventory = {
      hostname_type = "nic0_ip",
      host_vers = {
        nic1_ip = cidrhost("10.0.0.0/24", count.index + 1)
      }
    },
  })
  disks = [sakuracloud_disk.disk[count.index].id]

  network_interface {
    upstream = "shared"
  }
  network_interface {
    upstream = sakuracloud_switch.sw.id
  }

  user_data = templatefile("user-data.yaml", {
    fqdn = format("etcd%d", count.index + 1)
  })
}
