# Single CPP File Build

In this example we will show how we can wrap a simple CPP project with the default parameters, which will enable us to just provide a single CPP file.

If you want to wrap an existing library, or you have more complex requirements please refer to the ["Build system override CPP Wrapper example"](https://docs.seldon.io/projects/seldon-core/en/latest/examples/cpp_advanced.html).

You can read about how to configure your environment in the [CPP Wrapper page](https://docs.seldon.io/projects/seldon-core/en/latest/cpp/README.html).

## CPP Wrapper

The only thing that we will need for this, is a single CPP file. 

To ensure we align with the defaults, the name of our CPP file has to be `SeldonPackage.cpp`.

The contents of our file are quire minimal - namely:


```python
%%writefile SeldonPackage.cpp
#include "seldon/SeldonModel.hpp"

class ModelClass : public seldon::SeldonModelBase {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

SELDON_DEFAULT_BIND_MODULE()

```

    Overwriting SeldonPackage.cpp


In this file we basically have to note the following key points:

* We import `"seldon/SeldonModel.hpp"` which is from the Seldon package
* For the defaults the name of our class has to be `"ModelClass"`
* We extend the `SeldonModelBase` class which processes the protos for us
* We override the `predict()` function which provides the raw protos
* We register our class as `SELDON_DEFAULT_BIND_MODULE`

## Build Seldon Microservice

We can now build our seldon microservice using `s2i`:


```python
!s2i build . seldonio/s2i-cpp-build:0.0.1 seldonio/simple-cpp:0.1
```

    ---> Installing application source...
    ---> Installing application ...
    Looking in links: /whl
    Obtaining file:///microservice
    Installing collected packages: SeldonPackage
    Running setup.py develop for SeldonPackage
    Successfully installed SeldonPackage
    WARNING: Url '/whl' is ignored. It is either a non-existing path or lacks a specific scheme.
    WARNING: You are using pip version 20.2; however, version 22.0.4 is available.
    You should consider upgrading via the '/opt/conda/bin/python -m pip install --upgrade pip' command.
    Collecting pip-licenses
    Downloading pip_licenses-3.5.3-py3-none-any.whl (17 kB)
    Collecting PTable
    Downloading PTable-0.9.2.tar.gz (31 kB)
    Building wheels for collected packages: PTable
    Building wheel for PTable (setup.py): started
    Building wheel for PTable (setup.py): finished with status 'done'
    Created wheel for PTable: filename=PTable-0.9.2-py3-none-any.whl size=22908 sha256=3957aaca98c08e334c02711f973964454d089ed7ec1d25760986f01af67dde69
    Stored in directory: /root/.cache/pip/wheels/1b/3a/02/8d8da2bca2223dda2f827949c88b2d82dc85dccbc2bb6265e5
    Successfully built PTable
    Installing collected packages: PTable, pip-licenses
    Successfully installed PTable-0.9.2 pip-licenses-3.5.3
    WARNING: You are using pip version 20.2; however, version 22.0.4 is available.
    You should consider upgrading via the '/opt/conda/bin/python -m pip install --upgrade pip' command.
    created path: ./licenses/license_info.csv
    created path: ./licenses/license.txt
    Build completed successfully


## Test our model locally by running docker


```python
!docker run --name "simple_cpp" -d --rm -p 9000:9000 seldonio/simple-cpp:0.1
```

    8027cd995ae3012286abfccaa0cb12be0688997f2d156d0d415e118b06a3a60f


### Send request (which should return the same value)


```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '{"strData":"hello"}' \
    http://localhost:9000/api/v1.0/predictions
```

    {"strData":"hello"}

### Clean up


```python
!docker rm -f "simple_cpp"
```

    simple_cpp


## Deploy to seldon


```bash
%%bash
kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: simple-cpp
spec:
  predictors:
  - componentSpecs:
    - spec:
        containers:
          - image: seldonio/simple-cpp:0.1
            name: classifier
    engineResources: {}
    graph:
      name: classifier
      type: MODEL
    name: default
    replicas: 1
END
```

    seldondeployment.machinelearning.seldon.io/simple-cpp created



```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '{"strData":"hello"}' \
    http://localhost:80/seldon/default/simple-cpp/api/v1.0/predictions
```

    {"strData":"hello"}


```python
!kubectl delete sdep simple-cpp
```

    seldondeployment.machinelearning.seldon.io "simple-cpp" deleted



```python

```
