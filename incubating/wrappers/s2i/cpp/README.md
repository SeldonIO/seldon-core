# Seldon CPP Wrapper Module

This module conatains the base libraries and dependencies required for wrapping C++ code.

The dependencies for building are:

* protoc (with libprotoc) 3.0.0+
* Python (with dev libraries) 3.6+
* pybind11 (provided in this repo via makefile)

# To build and run tests

We provide a top level makefile which you can use to build all relevant files.

```
make cmake
make cmake-test
```

You are also able to just build all the relevant files using cmake - we advise using out of source builds (ie passing -bBUILDDIR) as otherwise the top level makefile may get overriden. 

