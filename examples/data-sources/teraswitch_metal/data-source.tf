data "teraswitch_metal" "example" {
  id = 12345
}

# Use the retrieved data
output "metal_ip_addresses" {
  value = data.teraswitch_metal.example.ip_addresses
}

output "metal_status" {
  value = data.teraswitch_metal.example.status
}
