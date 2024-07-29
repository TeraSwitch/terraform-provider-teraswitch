resource "teraswitch_volume" "my-volume" {
  region_id    = "PIT1"
  display_name = "yeehaw"
  size         = 20
  volume_type  = "nvme"
  description  = "My volume"
}
