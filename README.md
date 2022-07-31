[![codecov](https://codecov.io/gh/mdreem/s3_terraform_registry/branch/master/graph/badge.svg?token=33R19vS2kL)](https://codecov.io/gh/mdreem/s3_terraform_registry)

# s3_terraform_registry
A simple S3 backed terraform registry

# Usage

`s3_terraform_registry` needs a S3 bucket as a backend. The file structure in S3 is as follows:

For every combination of version, platform and architecture create a zip. Upload it into the respective
version folder:

```text
<namespace>/<type>/<version>/terraform-provider-<type>_<version>_<platform>_<architecture>.zip
```

Add a file `<namespace>/<type>/<version>/shasum` containing the sha256 sums of the zip-files. Also add
a file `<namespace>/<type>/<version>/shasum.sig`, which is the signature of the shasum-file.

Add the keyfile called `keyfile` which contains the public key used to create `shasum.sig` and put the key id
in the fil `key_id`.

At minimum the one provider version would consist of the following files in S3:

```text
<namespace>/<type>/<version>/<name>_<version>_<os>_<arch>.zip
<namespace>/<type>/<version>/shasum
<namespace>/<type>/<version>/shasum.sig
<namespace>/<type>/<version>/keyfile
<namespace>/<type>/<version>/key_id
```

`<name>` should not contain any underscore because the file will not be found in that case.

## Configuration

The registry is configured via the following flags:

- `bucket-name`: This is the S3 bucket where the files are placed.
- `hostname`: The hostname under which this registry will be available.
- `region`: Needs to be set to the region where the bucket resides in. E.g. eu-central-1.
- `port`: (optional) port the registry will listen on.
- `loglevel`: (optional) can be set to `error`, `info`, `debug` to set loglevel.
