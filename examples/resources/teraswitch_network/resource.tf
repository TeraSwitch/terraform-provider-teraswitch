resource "teraswitch_network" "my_network" {
  region_id      = "SLC1"
  display_name   = "my-private-network"
  v4_subnet      = "10.0.0.0"
  v4_subnet_mask = "255.255.255.0"
}
