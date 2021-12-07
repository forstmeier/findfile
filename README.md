# findfile

> API-first image file text search 🔍

## About

`findfile` is the root API implementation of the file search service.  

Store, query, and manage your JPGs, PNGs, and PDFs like you're searching text documents.  

## Setup

### Prerequisites

In order to work with the scripts in `bin`, you'll need to have the following installed:  

- [jq](https://stedolan.github.io/jq/) - version `jq-1.6`  
- [AWS CLI](https://aws.amazon.com/cli/) - version `aws-cli/1.19.53 Python/3.8.10 Linux/5.11.0-36-generic botocore/1.20.53`  

This code has been developed locally on an Ubuntu machine and has not been tested on other systems.  

### Installation

For quickstart run the following command and follow the prompts.  

```bash
bash <(curl -s https://raw.githubusercontent.com/forstmeier/findfile/master/bin/quickstart) | tee "quickstart-$(date +%Y%m%d-%H%M).log"  
```

For more in-depth usage and configuration, clone this repository, add an `etc/config/config.json` file (in the structure seen in the `bin/create_release` script), and run the scripts available in `bin`.  

## Usage

The `findfile` application listens to file events emitted by configured target S3 buckets. It then updates the database with that file data which can then be queried by the user. Two endpoints are provided:  

- `/buckets` is responsible for adding and removing target buckets
- `/documents` is responsible for running queries against the database

Below is an example `buckets` query to add and remove buckets.  

```bash
curl -X PUT https://7z8ruudxc9.execute-api.us-east-1.amazonaws.com/production/buckets --header "Content-Type: application/json" --header "x-findfile-security-key: 6758db58-9534-4e63-8eb9-ff402f6c29d7" --data '{"add": ["new-target-bucket"], "remove": ["old-target-bucket"]}'
```

Below is an example `documents` query searching for the text `"find me"`.  

```bash
curl -X PUT https://7z8ruudxc9.execute-api.us-east-1.amazonaws.com/production/documents --header "Content-Type: application/json" --header "x-findfile-security-key: 6758db58-9534-4e63-8eb9-ff402f6c29d7" --data '{"text": "find me"}'
```

A successful query response will contain the bucket and key values for any files matching the query text.  

### Notes

A couple of caveats and potential future changes to be aware of:

1. AWS does not currently support the correct event when deleting files through the S3 console for `findfile` to correctly listen to; if this is a significant issue, we can look into a solution.
2. [S3 event notifications](https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html) may be introduced to the current "listening" architecture (this would likely address the above issue).
3. The stack is not currently very configurable but it could be expanded going forward if needed.
4. Current database implementation defaults are in order to maintain a free tier option but these can be increased if there is interest.

## Contribute

Fork this repository and send a pull request. Follow Go best practices for structure and formatting! 