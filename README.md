# api

> API-first image file text search

## Description

`api` is the root API implementation of the **[FindFile](https://findfiledev.github.io)** service. 🔍  

## Setup

### Introduction

Several prerequisites are required for working on the `api` package code.  

- [Go](https://golang.org/dl/)
- [Git](https://git-scm.com/downloads)
- [AWS CLI](https://aws.amazon.com/cli/)

Only the **AWS CLI** is required in order to launch the API into the target AWS account.  

All infrastructure is defined in pure [AWS CloudFormation](https://aws.amazon.com/cloudformation/). The CloudFormation template (`cft.yaml`) assumes that two S3 buckets exist prior to launch - `ArtifactBucket` (storing application artifacts) and `DatabaseBucket` (storing application database data). Provide these bucket names to the template parameters whowever the stack is launched.  

Below is an overview of how to launch the `api` package.  

### Stack

The `api` CloudFormation stack can be created in two ways.  

- **Release package**
	- Download the most recent package from the [releases](https://github.com/findfiledev/api/releases) page and extract the resources
	- This contains the CloudFormation template (`cft.yaml`), three Lambda binaries (`database.zip`, `query.zip`, and `setup.zip`), a config file (`config.json`) and two helper scripts (`start_api`, and `add_source`)
	- Update `config.json` fields marked `"REPLACE"` with your pre-existing `ArtifactBucket` and `DatabaseBucket` names and optionally set the stack name field (defaulting to "findfile")
	- Run the `start_api` script with the required arguments to launch the stack
- **Helper scripts**
	- Add a `config.json` file to the existing `etc/config/` directory in this repository (see below for an example)
	- Change the permissions on the scripts to allow execution by running `chmod +x bin/`
	- Run the scripts in the `bin/` folder from the root directory (e.g. `./bin/build_lambdas`) to manually construct the stack in stages

> Example `config.json` file

```json
{
	"aws": {
		"cloudformation": {
			"stack_name": "findfile"
		},
		"s3": {
			"artifact_bucket": "findfile-artifact",
			"database_bucket": "findfile-database"
		}
	}
}
```

### Sources

Any S3 buckets containing image files are the data source that the `api` package consumes for the database - these are called **source buckets**.  

In order to be setup for the `api`, they require a [bucket policy](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucket-policies.html) and [event notifications](https://docs.aws.amazon.com/AmazonS3/latest/userguide/NotificationHowTo.html) to be configured. The recommended way to do this is to run the `add_source` script from the release package. The target source bucket bucket name is a required argument for the script (e.g. `./add_source new_source_bucket_name`).  

- **Note**: any existing bucket policy will be overwritten by this script  
- **Note**: there may be collisions with existing event notification configurations  
- **Note**: this script applies the event notifications to the full bucket not a prefix  
- **Note**: pre-existing files in the source bucket will not be added to the `api`; only files uploaded after launching the stack and configuring the bucket policy and event notifications will be added  

### Database

The S3 bucket added as the `DatabaseBucket` parameter in the stack creation holds the data queried by the `api` - this is called the **database bucket**.  

This should be a pre-existing bucket that will be retained despite the stack being torn down. As part of the stack creation a [bucket policy](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucket-policies.html) is placed on the `DatabaseBucket` to provide the required access for the `api`.  

- **Note**: the stack may overwrite any existing bucket policies on `DatabaseBucket`  
- **Note**: the role ARN is obfuscated by AWS in the bucket policy if the role is deleted as a safety precaution  

## Usage

There are three main steps to get up and running with the `api` package.  

1. Launch the stack using one of the options [listed above](###stack)  
2. Add source buckets following the [above instructios](###sources)  
3. Execute queries against the API endpoint using [FQL](###fql) statements  

### FQL

**FQL** is a basic query language used by the **FindFile** API and is represented in JSON format. Below is an example of a query payload.  

```json
{
	"search": {
		"text": "hello, world",
		"page_number": 1,
		"coordinates": [
			[0.0,0.0],
			[0.5,0.5]
		]
	}
}
```

The `"search"` object must contain:  

- `"text"`: a string value that the API will find matches to in the stored files  
- `"page_number"`: an integer indicating which page of the file to search on  
- `"coordinates"`: an array of two arrays containing floating point values between 0.0 and 1.0 which represent the top left and bottom right coordinates of the area of the page to search for text in (e.g. `[0.0,0.0]` is the top left corner of the page and `[1.0, 1.0]` is the bottom right corner of the page)  

A successful invocation will be a JSON object containing a `"message"` field indicating success and a `"data"` field holding an object containing bucket name keys and values of file name arrays.  

### Request

Below is a basic example of a request sent using [cURL](https://curl.se/). The security header is required to invoke the API endpoint. The API URL, security header, and security key values are all available as outputs in the CloudFormation stack as `QueryAPIEndpoint`, `HTTPSecurityKeyHeader`, and `HTTPSecurityKeyValue` respectively.  

```
curl -X POST https://<api_id>.execute-api.us-east-1.amazonaws.com/production/query --header "Content-Type: application/json" --header "x-findfile-security-key: <security_key>" --data '{"search": {"text": "hello, world", "page_number": 1, "coordinates": [[0.0,0.0], [0.5,0.5]]}}'
```

## Future

Some potential future expansions include:  

- **_Bulk file ingestion_** on adding a new source bucket  
- Providing **_multiple or nested FQL_** queries per request  
- **_TBD_**!  
