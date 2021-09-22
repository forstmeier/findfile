# api

## description

`api` is the root API implementation of the **[FindFile](https://findfiledev.github.io)** service. 🔍

## setup

Several prerequisites are required for working on the `api` package code.

- [Go](https://golang.org/dl/)
- [Git](https://git-scm.com/downloads)
- [AWS CLI](https://aws.amazon.com/cli/)

Only the **AWS CLI** is required in order to launch the API into the target AWS account.

All infrastructure is defined in pure [AWS CloudFormation](https://aws.amazon.com/cloudformation/). The CloudFormation template (`cft.yaml`) assumes that two S3 buckets exist prior to launch - `ArtifactBucket` (storing application artifacts) and `DatabaseBucket` (storing application database data). Provide these bucket names to the template parameters whowever the stack is launched.

## usage

Below is an overview of how to start using the `api` package.

### launch

The `api` CloudFormation stack can be created in two ways.

- **Release package**
	- Download the most recent package from the [releases](https://github.com/findfiledev/api/releases) page and extract the resources
	- This contains the CloudFormation template (`cft.yaml`), three Lambda binaries (`database`, `query`, and `setup`), and the starter script `start_api`
	- Run the `start_api` script with the required arguments to launch the stack
- **Helper scripts**
	- Add a `config.json` file to the existing `etc/config/` directory (see below for an example)
	- Change the permissions on the scripts to allow execution by running `chmod +x bin/`
	- Run the scripts in the `bin/` folder from the root directory (e.g. `./bin/build_lambdas`) to manually construct the stack in stages

> Example `config.json` file

```json
{
	"app": {
		"name": "findfile"
	},
	"aws": {
		"s3": {
			"artifact_bucket": "findfile-artifact",
			"database_bucket": "finefile-database"
		}
	}
}
```

### sources

S3 buckets containing image files are the data source the `api` package consumes and exposes for querying. Users can configure buckets for the `api` to listen on in two ways.

- **S3 console**: in the [S3 console](https://s3.console.aws.amazon.com/s3), on the **Permissions** tab for the target bucket, the user can manually add a JSON-structured **Bucket policy**
- **Helper script**: the `create_policy` script in the `bin/` folder can be run with the required arguments to programmatically apply a policy with the required permissions

**Note**: pre-existing files in a bucket are not added to the database; only files uploaded after launching the `api` stack and adding the bucket policy are indexed
**Note**: for both options, the required query Lambda role ARN is available in the CloudFormation stack outputs
**Note**: the role ARN is obfuscated by AWS in the bucket policy if the role is deleted as a safety precaution

## future

Some potential future expansions include:

- Bulk file ingestion on adding a new target bucket
- Providing multiple or nested FQL queries per request
- TBD
