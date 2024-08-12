# Configure the SakuraCloud Provider
terraform {
  required_providers {
    sakuracloud = {
      source = "sacloud/sakuracloud"

      # We recommend pinning to the specific version of the SakuraCloud Provider you're using
      # since new versions are released frequently
      version = "2.25.0"
      #version = "~> 2"
    }
  }
}
provider "sakuracloud" {
  # More information on the authentication methods supported by
  # the SakuraCloud Provider can be found here:
  # https://docs.usacloud.jp/terraform/provider/

  profile = "default"
}

resource "sakuracloud_switch" "is1b" {
  name        = "is1b"
  bridge_id   = sakuracloud_bridge.bridge.id
  zone        = "is1b"
}

resource "sakuracloud_switch" "tk1b" {
  name        = "tk1b"
  bridge_id   = sakuracloud_bridge.bridge.id
  zone        = "tk1b"
}

resource "sakuracloud_bridge" "bridge" {
  name        = "bridge"
}

resource "sakuracloud_server" "is1b_server" {
  name        = "is1b_server"
  disks       = [sakuracloud_disk.is1b_disk.id]

  network_interface {
    upstream         = "shared"
  }

  network_interface {
    upstream = sakuracloud_switch.is1b.id
  }

  user_data = templatefile("user-data.yaml", {
    fqdn = "is1b"
  })
  zone = "is1b"
}

resource "sakuracloud_server" "tk1b_server" {
  name        = "tk1b_server"
  disks       = [sakuracloud_disk.tk1b_disk.id]

  network_interface {
    upstream         = "shared"
  }

  network_interface {
    upstream = sakuracloud_switch.tk1b.id
  }

  user_data = templatefile("user-data.yaml", {
    fqdn = "tk1b"
  })
  zone = "tk1b"
}


data "sakuracloud_archive" "is1b_ubuntu" {
  zone = "is1b"
  filter {
    tags = [
      "cloud-init",
      "distro-ubuntu",
      "distro-ver-22.04.3",
    ]
  }
}

data "sakuracloud_archive" "tk1b_ubuntu" {
  zone = "tk1b"
  filter {
    tags = [
      "cloud-init",
      "distro-ubuntu",
      "distro-ver-22.04.3",
    ]
  }
}

resource "sakuracloud_disk" "is1b_disk" {
  name              = "is1b_disk"
  source_archive_id = data.sakuracloud_archive.is1b_ubuntu.id
  zone = "is1b"
}

resource "sakuracloud_disk" "tk1b_disk" {
  name              = "tk1b_disk"
  source_archive_id = data.sakuracloud_archive.tk1b_ubuntu.id
  zone = "tk1b"
}