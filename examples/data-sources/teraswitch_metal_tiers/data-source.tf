data "teraswitch_metal_tiers" "all" {}

# Output all available metal tiers
output "available_tiers" {
  value = data.teraswitch_metal_tiers.all.tiers
}

# Output tier IDs for easy reference
output "tier_ids" {
  value = [for tier in data.teraswitch_metal_tiers.all.tiers : tier.id]
}

# Example: Find a specific tier by CPU
output "amd_epyc_tiers" {
  value = [for tier in data.teraswitch_metal_tiers.all.tiers : tier if can(regex("EPYC", tier.cpu))]
}
