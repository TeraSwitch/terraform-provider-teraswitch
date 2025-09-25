# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.4] - 2025-01-15

### Changed

- **BREAKING**: Updated minimum Go version requirement from 1.21 to 1.24
- Updated Terraform Plugin Framework from v1.12.0 to v1.16.0
- Updated Terraform Plugin Framework Validators from v0.14.0 to v0.18.0
- Updated Terraform Plugin Go from v0.24.0 to v0.29.0
- Updated Terraform Plugin Testing from v1.10.0 to v1.13.3
- Updated Terraform Plugin Docs from v0.19.4 to v0.23.0
- Updated oapi-codegen from v2.4.1 to v2.5.0

### Fixed

- Improved error handling in HTTP response body cleanup due to package changes
- Fixed unused variable assignment in volume resource due to package changes

## [0.0.3] - 2024-07-28

### Added

- Enhanced documentation for all resources (cloud_compute, metal, network, volume)
- Added comprehensive examples for all resource types
- Added example configurations in `examples/resources/` directory

### Changed

- Updated provider documentation with better usage examples
- Improved resource documentation with detailed attribute descriptions
- Updated network resource configuration

### Fixed

- Improved resource examples and documentation clarity

## [0.0.2] - 2024-07-28

### Fixed

- Fixed GitHub Actions release workflow configuration
- Corrected release action permissions and settings

## [0.0.1] - 2024-07-28

### Added

- Initial release of Terraform Provider for Teraswitch
- Support for managing Teraswitch cloud resources
- **Resources:**
  - `teraswitch_cloud_compute` - Manage cloud compute instances
  - `teraswitch_metal` - Manage bare metal servers with power state control
  - `teraswitch_network` - Manage network resources
  - `teraswitch_volume` - Manage storage volumes
- **Provider Configuration:**
  - API key authentication
  - Project ID specification
- **Features:**
  - Complete CRUD operations for all resources
  - Import functionality for existing resources
  - Comprehensive validation and error handling
  - Generated client code using oapi-codegen
- **Development Tools:**
  - Automated testing with GitHub Actions
  - Dependabot for dependency updates
  - golangci-lint for code quality
  - Terraform Plugin Framework for modern provider development

### Requirements

- Terraform >= 1.0
- Go >= 1.21 (initial release)
