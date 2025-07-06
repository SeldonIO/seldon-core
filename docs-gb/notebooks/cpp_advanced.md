# Advanced CPP Buildsystem Override

In this example we will show how we can wrap a complex CPP project by extending the buildsystem defaults provided, which will give us flexibility to configure the required bindings.

If you are looking for a basic implementation of the C++ wrapper, you can get started with the ["Single file C++ Example"](https://docs.seldon.io/projects/seldon-core/en/latest/examples/cpp_simple.html).

You can read about how to configure your environment in the [CPP Wrapper documentation page](https://docs.seldon.io/projects/seldon-core/en/latest/cpp/README.html).

## Naming Conventions

In this example we will have full control on naming conventions.

More specifically there are a few key naming conventions that we need to consider:
* Python Module name
* Python Wrapper Class name
* C++ Library Name

As long as we keep these three key naming conventions in mind, we will have full flexibility on the entire build system.

For this project we will choose the following naming conventions:
* Python Module Name: `CustomSeldonPackage`
* Python Wrapper Class: `MyModelClass`
* C++ Library Name: `CustomSeldonPackage`

As you can see, the name of the Python Module and C++ Library can be the same.

## Wrapper Class

We will first start with the wrapper code of our example. We'll first create our file `Main.cpp` and we'll explain in detail each section below.


```python
%%writefile Main.cpp
#include "seldon/SeldonModel.hpp"

class MyModelClass : public seldon::SeldonModelBase {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

SELDON_BIND_MODULE(CustomSeldonPackage, MyModelClass)

```

    Overwriting Main.cpp


In this file we basically have to note the following key points:

* We import `"seldon/SeldonModel.hpp"` which is from the Seldon package
* We use our custom class name `"MyModelClass"`
* We extend the `SeldonModelBase` class which processes the protos for us
* We override the `predict()` function which provides the raw protos
* We register our class as `SELDON_BIND_MODULE` passing the package name and class name

## Buildsystem CMakeLists.txt
For the build system we have integrated with CMake, as this provides quite a lot of flexibility, and easy integration with external projects.

In this case below are the minimal configurations required in order for everything to work smoothly. The key components to note are:

* We fetch the seldon and pybind11 packages
* We register our C++ library with the name `CustomSeldonMessage`
* We bind the package with the seldon library

You are able to extend the points below as required.


```python
%%writefile CMakeLists.txt
cmake_minimum_required(VERSION 3.4.1)
project(seldon_custom_model VERSION 0.0.1)

set(CMAKE_CXX_STANDARD 14)

find_package(seldon REQUIRED)
find_package(pybind11 REQUIRED)

pybind11_add_module(
    CustomSeldonPackage
    Main.cpp)

target_link_libraries(
    CustomSeldonPackage PRIVATE
    seldon::seldon)
```

    Overwriting CMakeLists.txt


# Environment Variables
The final component is to specify the environment variables. 

FOr this we can either pass the env variable as a parameter to the `s2i` command below, or in this example we'll approach it by the other option which is creating an environment file in the `.s2i/environment` file.

The environment variable is `MODEL_NAME`, which should contain the name of your package and model. 

In our case it is `CustomSeldonPackage.MyModelClass` as follows:


```python
!mkdir -p .s2i/
```


```python
%writefile .s2i/environment
MODEL_NAME = CustomSeldonPackage.MyModelClass
```

    UsageError: Line magic function `%writefile` not found (But cell magic `%%writefile` exists, did you mean that instead?).


## (Optional) Extend CMake Config via Setup.py

In our case we won't have to pass any custom CMAKE parameters as we can configure everything through the `CMakeLists.txt`, but if you wish to modify how your C++ wrapper is packaged you can extend the setup.py file by following the details in the CPP Wrapper documentation page.

## Build Seldon Microservice

We can now build our seldon microservice using `s2i`:


```python
!s2i build . seldonio/s2i-cpp-build:0.0.1 seldonio/advanced-cpp:0.1
```

    ---> Installing application source...
    ---> Installing application ...
    Looking in links: /whl
    Obtaining file:///microservice
    Installing collected packages: SeldonPackage
    Running setup.py develop for SeldonPackage
    Successfully installed SeldonPackage
    Collecting pip-licenses
    Downloading https://files.pythonhosted.org/packages/08/b4/6e53ab4e82e2b9f8356dd17d7b9e30cba57ba0460186c92cc05e8a1a7f97/pip_licenses-3.0.0-py3-none-any.whl
    Collecting PTable (from pip-licenses)
    Downloading https://files.pythonhosted.org/packages/ab/b3/b54301811173ca94119eb474634f120a49cd370f257d1aae5a4abaf12729/PTable-0.9.2.tar.gz
    Building wheels for collected packages: PTable
    Building wheel for PTable (setup.py): started
    Building wheel for PTable (setup.py): finished with status 'done'
    Created wheel for PTable: filename=PTable-0.9.2-cp37-none-any.whl size=22906 sha256=98facc4ac39cd0e7c89a7c87587cf9941e9aa75817f105b8e5e01b499d1efb2a
    Stored in directory: /root/.cache/pip/wheels/22/cc/2e/55980bfe86393df3e9896146a01f6802978d09d7ebcba5ea56
    Successfully built PTable
    Installing collected packages: PTable, pip-licenses
    Successfully installed PTable-0.9.2 pip-licenses-3.0.0
    created path: ./licenses/license_info.csv
    created path: ./licenses/license.txt
    Build completed successfully


## Test our model locally by running docker


```python
!docker run --name "advanced_cpp" -d --rm -p 9000:9000 seldonio/advanced-cpp:0.1
```

    aaa5795779f2e605f7ead2772e912c8dd7de04002457eb4b3966b2b2182c63f4


### Send request (which should return the same value)


```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '{"strData":"hello"}' \
    http://localhost:9000/api/v1.0/predictions
```

    {"strData":"hello"}

### Clean up


```python
!docker rm -f "advanced_cpp"
```

    advanced_cpp


## Deploy to seldon


```bash
%%bash
kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: advanced-cpp
spec:
  predictors:
  - componentSpecs:
    - spec:
        containers:
          - image: seldonio/advanced-cpp:0.1
            name: classifier
    engineResources: {}
    graph:
      name: classifier
      type: MODEL
    name: default
    replicas: 1
END
```

    seldondeployment.machinelearning.seldon.io/advanced-cpp created



```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '{"strData":"hello"}' \
    http://localhost:80/seldon/default/advanced-cpp/api/v1.0/predictions
```

    {"strData":"hello"}


```python
!kubectl delete sdep advanced-cpp
```

    seldondeployment.machinelearning.seldon.io "advanced-cpp" deleted



```python

```
