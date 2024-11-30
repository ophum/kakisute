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
  zone    = "tk1b"
}


