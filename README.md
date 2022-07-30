[![codecov](https://codecov.io/gh/mdreem/s3_terraform_registry/branch/master/graph/badge.svg?token=33R19vS2kL)](https://codecov.io/gh/mdreem/s3_terraform_registry)

# s3_terraform_registry
A simple S3 backed terraform registry

# Usage

`s3_terraform_registry` needs a S3 bucket as a backend. The file structure is as follows:

For every binary there is an additional `shasum` and a `shasum.sig` which signs the `shasum`.
This signature can be checked with the public key that has to be given.
```text
<namespace>/<type>/<version>/<name>_<version>_<os>_<arch>.zip
<namespace>/<type>/<version>/shasum
<namespace>/<type>/<version>/shasum.sig
<namespace>/<type>/<version>/key_id
<namespace>/<type>/<version>/keyfile
```

`<name>` should not contain any underscore because the file will not be found in that case.

## Configuration

The registry is configured via the following flags:

- `bucket-name`: This is the S3 bucket where the files are placed.
- `hostname`: The hostname under which this registry will be available.
- `port`: (optional) port the registry will listen on.
- `loglevel`: (optional) can be set to `error`, `info`, `debug` to set loglevel.

# File structure in S3

For every combination of version, platform and architecture create a zip. Upload it into the respective
version folder:

```text
<namespace>/<type>/<version>/terraform-provider-<type>_<version>_<platform>_<architecture>.zip
```

Add a file `<namespace>/<type>/<version>/shasum` containing the sha256 sums of the zip-files. Also add
a file `<namespace>/<type>/<version>/shasum.sig`, which is the signature of the shasum-file. 
