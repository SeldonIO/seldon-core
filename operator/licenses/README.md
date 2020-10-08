# Operator licenses

## Usage

Use the `Makefile` in the parent `operator` folder:

```bash
make install-dev
make licenses
```

## How it works

Follow https://github.com/kubeflow/testing/tree/master/py/kubeflow/testing/go-license-tools

e.g.,

get-github-repo
get-github-license-info --github-api-token-file=/home/clive/.github_api_token
python ~/work/kubeflow/testing/py/kubeflow/testing/go-license-tools/patch_additional_license_info.py
concatenate-license
