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

### Step 3 - Add your source code

The core interface between your C++ code and Seldon is a wrapper class with the respective interface functions such as `predict`.

To simplify the interaction with Seldon, we provide a C++ module that contains relevant utilities, such as a parent wrapper class to inherit from, as well as the protobuf types.

The examples above provide insights on the structure of the classes that can be created. Below is a sample class with all the core elements.

```cpp
#include "seldon/SeldonModel.hpp"

class ModelClass : public seldon::SeldonModelBase {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

SELDON_DEFAULT_BIND_MODULE()
```

We'll break down each of the sections and provide further insights on the extensions and options available as alternatives.

#### Seldon C++ module include

The main include of the file `seldon/SeldonModel.hpp` provides a set of core utilities.

```cpp
#include "seldon/SeldonModel.hpp"
```

The core components that are imported are:

* The SeldonModel<proto> and SeldonModelBase class
* The SELDON_{X}_BIND_MODULE(...) macros
* The seldon::protos::SeldonMessage

#### SeldonModel class base

The next component is the SeldonModel class. As you saw in the sample above, we inherited the class as follows:

```cpp
class ModelClass : public seldon::SeldonModelBase {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};
```

The SeldonModel class provides two key components:

* It provides a `public virtual abstract` function `predict(proto)` which users are able to overide to add their custom logic
* Under the hood, SeldonModel also implements a `public virtual` method `predictRaw(py::bytes)` which basically receives the raw bytes, converts them into the relevant proto and passes it to the `predict(proto)` function

It's worth mentioning that SeldonModel<proto> is actually a template class, which enables for any protos to be provided. This of course is restricted through the service orchestrator but provides further flexibility.

More specifically, the SeldonModelBase class we use above is actually a template implementation as `using SeldonModelBase = SeldonModel<seldon::protos::SeldonMessage>;`.

#### BIND Macro

Finally we have the last step which is our binding macro. This is what tells Seldon to use our class provided above. By default, Selon expects the naming conventions `ModelClass` for the name of the class, and `SeldonPackage` for the name of the package itself.

Above we used the default binding, which was the following line:

```cpp
SELDON_DEFAULT_BIND_MODULE()
```

The macro above is equivallent to defining the following macro:

```cpp
SELDON_BIND_MODULE(SeldonPackage, ModelClass)
```

If you change either of the names, you will need to make sure you do the relevant overrides. To be more specific the changes required are below.

If you change the name of your ModelClass:

* Register it with the macro SELDON_BIND_MODULE
* Specify it in the MODEL env var to your s2i params (more on this in the optional steps below)

If you change the name of your SeldonPackage:
* Register it with the macro SELDON_BIND_MODULE
* Specify it in the MODEL env var to your s2i params (more on this in the next step)
* Override the package name in the buildsystem  (more on this in the optional steps below)

## Optional Steps

The steps above are optional, and are only required if you need the more advanced functionality, or alternatively if you want to change the naming convention.

### Step 4 (Optional) - Specify your seldon env variables

In order to specify your seldon environment variables, you are able to do so by one of the following:

* Provide them inside the `.s2i/environment` file where you run your `s2i` component
* Pass them as parameter values `-E` to your `s2i` command

The main variable that you may override would be:

* MODEL=SeldonPackage.ModelClass

As you can assume, this is the env variable that you'll need to make sure is set correctly for the Seldon wrapper to find your C++ package/class.

### Step 5 (Optional) - Overried Buildsystem

Finally you are also able to overide the buildsystem if you wish to do so for advanced configurations.

The build system we use is CMAKE, primarily as it enables for flexible and modular development.

This means that you will be able to add a CMakeLists.txt file to your project folder. 

The contents of the file would be the following:

```
cmake_minimum_required(VERSION 3.4.1)
project(seldon_custom_model VERSION 0.0.1)

set(CMAKE_CXX_STANDARD 14)

find_package(seldon REQUIRED)
find_package(pybind11 REQUIRED)

pybind11_add_module(
    SeldonPackage
    ModelClass.cpp)

target_link_libraries(
    CustomSeldonPackage PRIVATE
    seldon::seldon)
```

The outline above is the simplest cmake file that you could create. It has the following components:

* find_package(seldon,...) - This fetches the seldon package
* find_package(pybind11, ...) - This fetches the requried bindings
* pybind11_add_module(SeldonPackage, ModelClass.cpp) - This creates your module with the name of your package, and all the relevant C++ source files.
* target_link_libraries(CustomSeldonPackage PRIVATE seldon::seldon) binds the Seldon shared/static library

Beyond this you are able to configure anything you wish for your environment. 

Further requirements can also be set by extending the docker imagebase, which is found in the module repo `incubating/wrappers/s2i/cpp/`.





