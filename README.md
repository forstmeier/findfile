# findfile

> API-first image file text search 🔍

## About

`findfile` is the root API implementation of the file search service.  

Store, query, and manage your JPGs, PNGs, and PDFs like you're searching text documents.  

## Setup

### Prerequisites

Several packages are required for launching and managing the `findfile` stack.  

- [jq](https://stedolan.github.io/jq/) - version `jq-1.6`  
- [AWS CLI](https://aws.amazon.com/cli/) - version `aws-cli/1.19.53 Python/3.8.10 Linux/5.11.0-36-generic botocore/1.20.53`  

### Installation

Follow the steps below to configure the required CloudFormation resources in your AWS account.  

- Download the most recent `release.zip` file from the [releases](https://github.com/forstmeier/findfile/releases) page  
- Extract the contents below into your desired folder
	- `cft.yaml`: the full CloudFormation template definition for the required AWS resources  
	- `setup.zip`, `database.zip`, `query.zip`: three AWS Lambda binaries pre-compiled and zipped  
	- `config.json`: configuration file with user-provided or generated information  
	- `start_api`: a Bash script file used to launch the CloudFormation stack  
	- `add_bucket`: a Bash script for adding user [S3 source buckets](####source-buket) to be listened to by the `findfile` service  
		<!-- ADD NOTES FROM EXISTING README REGARDING POLICY OVERWRITES -->
	- `generate_query`: a Bash script for generating cURL requests with correctly formatted FQL  
- All Bash scripts are used to manage the `findfile` service and reference the `config.json` which should live in the same directory  

### Terminology

#### Database bucket

An S3 bucket in the user's AWS account that stores the data used to answer user queries. This bucket is fully managed by the `findfile` service.  

> Do not edit this bucket  

#### Source bucket

An S3 bucket added to the `findfile` listener by the user that holds the raw image files the user wants to query via the API. These are configured by running the `add_bucket` script provided in the `release.zip` package file.  

#### Artifact bucket

An S3 bucket that stores the CloudFormation template (`cft.yaml`) and the AWS Lambda source code files (`setup.zip`, `database.zip`, and `query.zip`).  

> Do not edit this bucket  

#### FQL

A "query language" for querying the `findfile` database via the API. It is JSON-formatted and sent in an HTTPS `POST` request body.  Right now the FQL statement contains three parts:  

- `"text"`: a string value that the API will find matches to in the stored files  
- `"page_number"`: an integer indicating which page of the file to search on  
- `"coordinates"`: an array of two arrays containing floating point values between 0.0 and 1.0 which represent the top left and bottom right coordinates of the area of the page to search for text in (e.g. `[0.0,0.0]` is the top left corner of the page and `[1.0, 1.0]` is the bottom right corner of the page)  

Below is an example full FQL request object.  

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

## Usage

Follow the steps below to launch, configure, and interact with the `findfile` API.

1. Navigate to the folder where you downloaded and extracted the `release.zip` file in the [installation](###installation) step  
2. Update the `config.json` file with the required S3 bucket names  
	a. Option 1: update the `artifact_bucket` and `database_bucket` with pre-existing S3 buckets - this will allow `findfile` to upload artifacts and establish the database using the user-provided S3 buckets [1] [2]  
	b. Option 2: run `start_api` with the argument `create_buckets` (e.g. `./start_api create_buckets`) and `findfile` will create S3 buckets in the users AWS account [2]  
3. Run the `start_api` script which will upload the AWS Lambda source code zipped files and launch the full CloudFormation stack  
	a. This script will optionally create **database** and **artifact** S3 buckets if the `create_buckets` argument is provided  
4. Run the `add_bucket` script to add the desired S3 buckets containing image files to the `findfile` listener [3]  
	a. This script should be run for each S3 bucket the user wants to add (e.g. `./add_bucket bucket_name`)  
	b. The required [bucket policy](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucket-policies.html) and [event notifications](https://docs.aws.amazon.com/AmazonS3/latest/userguide/NotificationHowTo.html) are added by this script [4] [5] [6]  
5. Run the `generate_query` script with the user-provided arguments in order to generate fully-formed and ready-to-use- cURL commands to run agains the `findfile` API  
	a. First argument - _text_: a line of text enclosed in double quotes, example `"search for this text"`  
	b. Second argument - _page number_: an integer, example `1`  
	c. Third argument - _coordinates_: box quotes containing two box quotes enclosed in double quotes, example `"[[0.0, 0.0], [0.5,0.5]]"`  

An example series of commands:

```bash
>>> cd <path/to/download>
>>> ./start_api create_buckets
>>> ./add_buckets <your_source_bucket_name>
>>> ./generate_query "search for this text" 1 "[[0.0, 0.0], [0.5,0.5]]"
>>> curl -X POST <generated_url> --header "Content-Type: application/json" --header "x-findfile-security-key: <security_key>" --data '{"search": {"text": "search for this text", "page_number": 1, "coordinates": [[0.0,0.0], [0.5,0.5]]}}'
```

### Notes

[1] If the user provides pre-existing S3 buckets, the `findfile` stack will overwrite bucket policy configurations on the **database** bucket  
[2] The **database** and **artifact** buckets will not be deleted when the `findfile` stack is deleted to avoid accidental data loss  
[3] Files that exist in the **source** bucket prior to being added to the `findfile` listener via the `add_bucket` script will not be pulled into the database - only newly added or re-uploaded files will be added to the database for querying  
[4] The `add_bucket` script will apply a bucket policy that will overwrite existing configurations  
[5] If there are existing event notifications configured on the target **source** bucket, there may be collisions on events and prefixes  
[6] The applied event notifiaction does not currently include any prefixes  

## Roadmap

Features will be added according to overall project interest. Some potential future expansions include:  

- **_Bulk file ingestion_** on adding a new source bucket  
- **_Upgraded database_** throughput and querying features  
- Providing **_multiple or nested FQL_** queries per request  
- **_TBD_**!

## Contribute

There are a few tools required to begin working on the `findfile` codebase. The indicated versions are what the application was built using - other versions or operating systems have not been tested. See the contributing and code of conduct resources for specifics.  

- [Go](https://golang.org/dl/) - version `go version go1.16 linux/amd64`  
- [Git](https://git-scm.com/downloads) - version `git version 2.25.1`  
- [jq](https://stedolan.github.io/jq/) - version `jq-1.6`  
- [AWS CLI](https://aws.amazon.com/cli/) - version `aws-cli/1.19.53 Python/3.8.10 Linux/5.11.0-36-generic botocore/1.20.53`  

Scripts stored in the `bin/` folder are typically used for working with the `findfile` stack during development. A `config.json` file needs to be added at `etc/config/config.json` with user-provided pre-existing S3 buckets added to the respective `"REPLACE"` field values.  

```json
{
	"aws": {
		"cloudformation": {
			"stack_name": "findfile"
		},
		"s3": {
			"artifact_bucket": "REPLACE",
			"database_bucket": "REPLACE"
		}
	}
}
```
