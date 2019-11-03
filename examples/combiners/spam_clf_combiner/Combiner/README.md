# **Combiner**



#### Activate Virtual Env

Create a Python virtual environment:

```
    $ python3 -m venv venv
    $ source venv/bin/activate
```

#### Install s2i


    linux: https://computingforgeeks.com/install-source-to-image-toolkit-on-linux/
    mac:  brew install source-to-image


#### Build

```
$ s2i build . seldonio/seldon-core-s2i-python3:0.7 sandhya1594/Combiner:1.0.0.1
$ docker push sandhya1594/Combiner:1.0.0.1
```

#### Test

```
$ docker run --name "Combiner" --rm sandhya1594/Combiner:1.0.0.1

```