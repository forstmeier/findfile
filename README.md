# api

## description

`api` is the root API implementation of the [FindFile](https://findfile.dev) service. 🔍

## setup

Several prerequisites are required for working with the `api` package.

- [Go](https://golang.org/dl/)
- [Git](https://git-scm.com/downloads)
- [AWS CLI](https://aws.amazon.com/cli/)

## infrastructure

All infrastructure is defined in pure [AWS CloudFormation](https://aws.amazon.com/cloudformation/). The `bin` directory contains scrips for interacting with the infrastructure.

The CloudFormation template (`cft.yaml`) assumes that two S3 buckets exist prior to launch - `ArtifactBucket` (storing application artifacts) and `StorageBucket` (storing application data).

Values are passed into the template as parameters taken from the `config.json` file.
