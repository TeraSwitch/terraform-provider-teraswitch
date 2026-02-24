resource "teraswitch_ssh_key" "my_key" {
  display_name = "my-ssh-key"
  key          = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG... user@example.com"
}

# Reference the SSH key in other resources
resource "teraswitch_cloud_compute" "example" {
  project_id   = 123
  region_id    = "PIT1"
  tier_id      = "s1.1c1g"
  image_id     = "ubuntu-noble"
  display_name = "my-vm"
  ssh_key_ids  = [teraswitch_ssh_key.my_key.id]
}
