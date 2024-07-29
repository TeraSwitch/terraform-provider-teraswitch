resource "teraswitch_cloud_compute" "my-vm" {
  project_id          = 480
  region_id           = "PIT1"
  tier_id             = "s1.1c1g"
  image_id            = "ubuntu-noble"
  display_name        = "terraform-vm"
  ssh_key_ids         = [588]
  password            = null
  boot_size           = 64
  user_data           = null
  tags                = ["tag1", "tag2"]
  desired_power_state = null
  skip_wait_for_ready = false
}
