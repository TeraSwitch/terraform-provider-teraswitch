# Terraform Provider for TeraSwitch

## Using the provider

The Terraform TeraSwitch provider is a plugin that allows Terraform to manage
resources on [TeraSwitch](https://beta.tsw.io).

To use, add the following to your Terraform configuration:

```hcl
terraform {
  required_providers {
    teraswitch = {
      source = "TeraSwitch/teraswitch"
      version = "~> 0.0.9"
    }
  }
}

provider "teraswitch" {
    api_key    = "your-api-key"
    project_id = 123
}
```

## Resources and Data Sources

The provider supports the following resources and data sources:

### Resources
- `teraswitch_cloud_compute` - Manage cloud compute instances
- `teraswitch_metal` - Manage bare metal servers (now with import support!)
- `teraswitch_network` - Manage network resources
- `teraswitch_volume` - Manage storage volumes
- `teraswitch_ssh_key` - Manage SSH keys

### Data Sources
- `teraswitch_metal` - Query existing metal servers
- `teraswitch_ssh_keys` - Query all SSH keys in a project
- `teraswitch_metal_tiers` - Query available metal tiers with pricing
- `teraswitch_regions` - Query available regions with service type filtering

### Example: Using the Metal Data Source
```hcl
data "teraswitch_metal" "existing_server" {
  id = 12345
}

output "server_ip_addresses" {
  value = data.teraswitch_metal.existing_server.ip_addresses
}
```

### Example: Managing SSH Keys
```hcl
# Create a new SSH key
resource "teraswitch_ssh_key" "my_key" {
  display_name = "my-ssh-key"
  key          = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG... user@example.com"
}

# List all existing SSH keys
data "teraswitch_ssh_keys" "all" {}

# Use an SSH key in a compute resource
resource "teraswitch_cloud_compute" "example" {
  project_id   = 123
  region_id    = "PIT1"
  tier_id      = "s1.1c1g"
  image_id     = "ubuntu-noble"
  display_name = "my-vm"
  ssh_key_ids  = [teraswitch_ssh_key.my_key.id]
}
```

### Example: Querying Metal Tiers
```hcl
data "teraswitch_metal_tiers" "all" {}

output "available_tiers" {
  value = data.teraswitch_metal_tiers.all.tiers
}

# Get just the tier IDs
output "tier_ids" {
  value = [for tier in data.teraswitch_metal_tiers.all.tiers : tier.id]
}
```

### Example: Querying Regions
```hcl
# Get all regions
data "teraswitch_regions" "all" {}

# Get only regions that support Metal
data "teraswitch_regions" "metal" {
  service_type = "Metal"
}

output "metal_region_ids" {
  value = [for r in data.teraswitch_regions.metal.regions : r.id]
}
```

## What's New in v0.0.9

- **NEW**: Added `teraswitch_ssh_key` resource for managing SSH keys
- **NEW**: Added `teraswitch_ssh_keys` data source for querying all SSH keys in a project
- **NEW**: Added `teraswitch_metal_tiers` data source for querying available metal tiers with pricing
- **NEW**: Added `teraswitch_regions` data source for querying available regions with service type filtering
- Updated API client with new endpoints for SSH keys, regions, and tags

## Developing the Provider

### Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

To compile the provider, run `go install`. This will build the provider and put
the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`. Ensure
the following environment variables are set:

- `TERASWITCH_API_KEY`: API key for [beta.tsw.io](https://beta.tsw.io)
- `TERASWITCH_PROJECT_ID`: The ID of the project to create test resources in

_Note:_ Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

### Testing Against Dev Environment

For internal development and testing, you can configure the provider to use the dev API:

1. **Set up dev overrides** in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "TeraSwitch/teraswitch" = "/path/to/your/go/bin"
  }
  direct {}
}
```

2. **Set the required environment variables**:

```shell
export TERASWITCH_DEV_API_URL="http://your-dev-api-url"
export TERASWITCH_API_KEY="your-dev-api-key"
export TERASWITCH_PROJECT_ID="your-dev-project-id"
```

3. **Build and install the provider**:

```shell
go install
```

4. **Run Terraform** (skip `terraform init` when using dev overrides):

```shell
terraform plan
```

You should see a warning indicating that provider development overrides are in effect.

When `TERASWITCH_DEV_API_URL` is set, the provider will use that URL instead of the production API.
