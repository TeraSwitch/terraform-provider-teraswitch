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
      version = "~> 0.0.5"
    }
  }
}

provider "teraswitch" {
    api_key    = "your-api-key"
    project_id = 123
}
```

## Recent Updates

**v0.0.5** (Upcoming)

- Updated to Go 1.24
- Updated Terraform Plugin Framework to v1.16.0
- Updated Terraform Plugin Framework Validators to v0.18.0
- Updated all Terraform plugin dependencies for enhanced compatibility

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
