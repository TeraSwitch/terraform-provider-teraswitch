# Get all regions
data "teraswitch_regions" "all" {}

# Get only regions that support Metal
data "teraswitch_regions" "metal" {
  service_type = "Metal"
}

# Output all available regions
output "available_regions" {
  value = data.teraswitch_regions.all.regions
}

# Output region IDs for easy reference
output "region_ids" {
  value = [for region in data.teraswitch_regions.all.regions : region.id]
}

# Output regions that support Metal
output "metal_regions" {
  value = data.teraswitch_regions.metal.regions
}
