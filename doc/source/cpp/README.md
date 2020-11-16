# Packaging a C++ Framework/Model for Seldon Core

In this guide we cover how you can wrap your CPP models using the Seldon CPP wrapper.

For a quick start you can try out the following two examples:

* [Simple CPP Single File Example](../examples/cpp_simple)
* [Advanced CPP Build System Override Example](../examples/cpp_advanced)

The CPP Wrapper leverages the [Python Inference Server](../python) together with C++ bindings to communicate to the core CPP components natively.

If you are not familiar with s2i you can read [general instructions on using s2i](../wrappers/s2i.md) and then follow the steps below.

## Step 1 - Install s2i

[Download and install s2i](https://github.com/openshift/source-to-image#installation)

 * Prerequisites for using s2i are:
   * Docker
   * Git (if building from a remote git repo)

To check everything is working you can run

```bash
s2i usage seldonio/s2i-cpp-build:0.0.1
```

## Step 2 (Optional) - Install Seldon Core C++ Package

For this step you have two options:

* Install Seldon Core C++ package for easier development
* Don't install package but import subdirectory via your CMakeLists.txt file

### (Option 1) Install C++ Package

For the first option, you just have to go to `incubating/wrappers/s2i/wrappers/cpp/` and install using CMAKE.

As you would with cmake, you will be able to build the respective build components for your OS - for linux you can do:

```bash
cmake . -Bbuild

make

make install
```

Now in your CMakeLists.txt you are able to just add:

```
find_package(seldon REQUIRED)
```

### (Option 2) Import as subdirectory

With CMAKE you are able to import a subdirectory as project. For this you can just import using:

```
add_subdirectory(
    ../../../incubating/wrappers/s2i/cpp/
    ${CMAKE_CURRENT_BINARY_DIR}/seldon_dir})
```

Now you have imported the directory. If you do take this approach you need to remember that when you run the s2i command, this will perform the build in the image, which means that you should still add the find_package command under a if(...) guard.


The CPP wrapper is made possible thanks to the development in the Cython bindings and the [Pybind11 framework](https://github.com/pybind/pybind11), enabling for efficient and robust interoperability.

