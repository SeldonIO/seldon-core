# Release notes for master

## Changes


- Added `model.uri` and `model.implementation` keys in the `seldon-single-model` Helm chart. ([#2054](https://github.com/SeldonIO/seldon-core/pull/2054), [@RafalSkolasinski](https://github.com/RafalSkolasinski))
- Added `serviceAccountName` at the pod level to enable EKS fine-grained IAM roles. ([#1866](https://github.com/SeldonIO/seldon-core/pull/1866), [@enissay14](https://github.com/enissay14))
- Added custom error for R ([#2153](https://github.com/SeldonIO/seldon-core/pull/2153), [@johnny-butter](https://github.com/johnny-butter))
- Added gRPC support to model metadata. ([#2005](https://github.com/SeldonIO/seldon-core/pull/2005), [@RafalSkolasinski](https://github.com/RafalSkolasinski))
- Changed Azure dependency to be optional. ([#2170](https://github.com/SeldonIO/seldon-core/pull/2170), [@adriangonz](https://github.com/adriangonz))
- Changed Python wrapper so that it runs Gunicorn with a single worker by default.
  Added `GUNICORN_THREADS` environment variable to control the fixed number of threads on each Gunicorn worker. ([#2047](https://github.com/SeldonIO/seldon-core/pull/2047), [@adriangonz](https://github.com/adriangonz))
- Changed loggers in operator, executor and Python wrapper so that only `INFO` messages and above are logged by default.
  Added `SELDON_DEBUG` environment variable to operator, executor and Python wrapper to enable sampling and improve structured logging. ([#1980](https://github.com/SeldonIO/seldon-core/pull/1980), [@adriangonz](https://github.com/adriangonz))
- Fixed gRPC port name in executor. ([#2131](https://github.com/SeldonIO/seldon-core/pull/2131), [@groszewn](https://github.com/groszewn))
- Fixed semantics of `seldon.io/no-engine` annotation. ([#1970](https://github.com/SeldonIO/seldon-core/pull/1970), [@chengchengpei](https://github.com/chengchengpei))
- This will allow MLFlow_Server users to send the data.names parameter, and connect it with the DataFrame's columns property. Shouldn't break existing functionality, just extend it. ([#2135](https://github.com/SeldonIO/seldon-core/pull/2135), [@meoril](https://github.com/meoril))
- Update tracing dependencies for Python inference server. ([#2166](https://github.com/SeldonIO/seldon-core/pull/2166), [@adriangonz](https://github.com/adriangonz))
- Updated documentation about routing when using the Go executor. ([#2172](https://github.com/SeldonIO/seldon-core/pull/2172), [@cliveseldon](https://github.com/cliveseldon))
- Updated documentation to point to newer version of Ambassador. ([#2069](https://github.com/SeldonIO/seldon-core/pull/2069), [@RafalSkolasinski](https://github.com/RafalSkolasinski))
- Updated executor dependencies. ([#2099](https://github.com/SeldonIO/seldon-core/pull/2099), [@RafalSkolasinski](https://github.com/RafalSkolasinski))
- Updated operator dependencies. ([#2093](https://github.com/SeldonIO/seldon-core/pull/2093), [@RafalSkolasinski](https://github.com/RafalSkolasinski))
