resource "teraswitch_metal" "my-dedi" {
  region_id    = "SLC1"
  display_name = "terraform-metal"
  tier_id      = "7950x"
  project_id   = 480
  ssh_key_ids  = [588]
  password     = null
  tags         = ["tag1", "tag2"]
  memory_gb    = 128
  disks = {
    "nvme0n1" : "1.92t",
    "nvme1n1" : "1.92t",
  }
  raid_arrays = [
    {
      name        = "md0"
      type        = "Raid1"
      members     = ["nvme0n1", "nvme1n1"]
      file_system = "Ext4"
      mount_point = "/"
      size_bytes  = null
    },
  ]
  image_id            = "ubuntu-noble"
  reserve_pricing     = false
  desired_power_state = null
  wait_for_ready      = true
}
