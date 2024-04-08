terraform {
  backend "http" {
    address        = "http://localhost:8080/test"
    lock_address   = "http://localhost:8080/test"
    unlock_address = "http://localhost:8080/test"
  }
}

resource "local_file" "foo" {
  content  = "foo!"
  filename = "${path.module}/foo.bar"
}
