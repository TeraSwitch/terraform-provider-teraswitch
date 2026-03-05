# Get all tags in use across the project
data "teraswitch_tags" "all" {}

# Output all tags
output "all_tags" {
  value = data.teraswitch_tags.all.tags
}

# Example: Use tags to filter or organize resources
output "tag_count" {
  value = length(data.teraswitch_tags.all.tags)
}
