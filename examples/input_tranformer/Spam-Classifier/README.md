# **Spam Classifier**



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
$ s2i build . seldonio/seldon-core-s2i-python3:0.7 spam-classifier:1.0.0.1
$ docker push <yourdockerhubusername>/spam-classifier:1.0.0.1
```

#### Test

```
$ docker run --name "spam-classifier" --rm -d -p 5000:5000  spam-classifier:1.0.0.1

curl -g http://localhost:5000/predict --data-urlencode 'json={"data": {"names": ["message"], "ndarray": ["click here to win the price"]}}'

Result:

{"data":{"ndarray":["0.9785519250237192","spam"]},"meta":{}}

```




