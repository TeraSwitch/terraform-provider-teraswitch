# Terraform Provider for Teraswitch

## Using the provider

The Terraform Teraswitch provider is a plugin that allows Terraform to manage
resources on [Teraswitch](https://beta.tsw.io).

To use, add the following to your Terraform configuration:

```hcl
terraform {
  required_providers {
    teraswitch = {
      source = "TeraSwitch/teraswitch"
      version = "~> 0.0.7"
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

### Data Sources
- `teraswitch_metal` - Query existing metal servers (**NEW in v0.0.7**)

### Example: Using the Metal Data Source
```hcl
data "teraswitch_metal" "existing_server" {
  id = 12345
}

output "server_ip_addresses" {
  value = data.teraswitch_metal.existing_server.ip_addresses
}
```

## What's New in v0.0.7

- **NEW**: Added `teraswitch_metal` data source for querying existing metal servers
- **NEW**: Added import functionality for `teraswitch_metal` resources
- Enhanced metal resource documentation with import examples
- Improved test coverage and CI reliability

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
