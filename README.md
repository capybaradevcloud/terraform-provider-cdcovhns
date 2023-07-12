# CDC OVH Name Servers Provider

This repository contains a Terraform provider for managing name servers in OVH. The provider allows users to automate the configuration and management of Name Servers records within their OVH infrastructure using Terraform.

## Getting Started

To get started, follow the documentation: https://registry.terraform.io/providers/capybaradevcloud/cdcovhns/latest/docs

## TODO
- handle hosted name servers (if set to hosted, OVH will manage name servers and user cannot change it on its own)
- improve logging
- add name server status to tfstate (?)
- add Terraform version to Client User Agent
- add data sources
- add integration tests
- add more linters and improve CI/CD workflow
- improve error handling

## Contributing

Contributions to this provider are welcome! If you encounter any issues or have suggestions for improvements, please open an issue. Pull requests for new features, bug fixes, or documentation enhancements are also appreciated.

## Disclaimer

This provider is not officially maintained or supported by OVH. Please use it at your own risk and refer to the OVH documentation for official guidance on managing NS resources within OVH.
