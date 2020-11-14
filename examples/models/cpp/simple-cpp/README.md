# Single CPP File Build

In this example we will show how we can wrap a simple CPP project with the default parameters, which will enable us to just provide a single CPP file.

If you want to wrap an existing library, or you have more complex requirements please refer to the ["Build system override CPP Wrapper example"]().

You can read about how to configure your environment in the [CPP Wrapper page]().

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
    Collecting pip-licenses
    Downloading https://files.pythonhosted.org/packages/08/b4/6e53ab4e82e2b9f8356dd17d7b9e30cba57ba0460186c92cc05e8a1a7f97/pip_licenses-3.0.0-py3-none-any.whl
    Collecting PTable (from pip-licenses)
    Downloading https://files.pythonhosted.org/packages/ab/b3/b54301811173ca94119eb474634f120a49cd370f257d1aae5a4abaf12729/PTable-0.9.2.tar.gz
    Building wheels for collected packages: PTable
    Building wheel for PTable (setup.py): started
    Building wheel for PTable (setup.py): finished with status 'done'
    Created wheel for PTable: filename=PTable-0.9.2-cp37-none-any.whl size=22906 sha256=f8ceb5d135fae9aad1de576924d11560c8a004a59c489b980412a974eeffd694
    Stored in directory: /root/.cache/pip/wheels/22/cc/2e/55980bfe86393df3e9896146a01f6802978d09d7ebcba5ea56
    Successfully built PTable
    Installing collected packages: PTable, pip-licenses
    Successfully installed PTable-0.9.2 pip-licenses-3.0.0
    created path: ./licenses/license_info.csv
    created path: ./licenses/license.txt
    Build completed successfully


## Test our model locally by running docker


```python
!docker run --name "simple_cpp" -d --rm -p 5000:5000 seldonio/simple-cpp:0.1
```

    19f989a0ae389e041c4daf915067016a6ab8dc569873948a1099442a249ec502


### Send request (which should return the same value)


```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '{"strData":"hello"}' \
    http://localhost:5000/api/v1.0/predictions
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
