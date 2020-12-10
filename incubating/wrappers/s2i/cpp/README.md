# Seldon CPP Wrapper Module

This module conatains the base libraries and dependencies required for wrapping C++ code.

The dependencies for building are:

* protoc (with libprotoc) v3.14.0-rc2
* pybind11 (provided in this repo via makefile) v2.6.1
* Python (with dev libraries) 3.6+

# To build and run tests

We provide a top level makefile which you can use to build all relevant files.

```
make cmake
make cmake-test
```

You will notice that when running the tests we build with static library.

# Build Seldon Containers

In order to build for your deployments, you will want to make sure the library is installed as shared library, so it can be referenced in the build. You can see the examples in the examples folder.

