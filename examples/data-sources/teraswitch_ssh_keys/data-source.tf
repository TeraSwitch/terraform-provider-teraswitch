data "teraswitch_ssh_keys" "all" {}

# Output all SSH keys
output "all_ssh_keys" {
  value = data.teraswitch_ssh_keys.all.ssh_keys
}

# Use the first SSH key in a compute resource
resource "teraswitch_cloud_compute" "example" {
  project_id   = 123
  region_id    = "PIT1"
  tier_id      = "s1.1c1g"
  image_id     = "ubuntu-noble"
  display_name = "my-vm"
  ssh_key_ids  = [data.teraswitch_ssh_keys.all.ssh_keys[0].id]
}
