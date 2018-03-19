# Epsilon Greedy Router


## Wrap using s2i

```bash
s2i build . seldonio/seldon-core-s2i-python2 egreedy-router
```

## Smoke Test

Run under docker.

```bash
docker run --rm -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"name": "n_branches","value": "3","type": "INT"},{"name": "epsilon","value": "0.3","type": "FLOAT"},{"name": "verbose","value": "1","type": "BOOL"}]' egreedy-router
```

Send a data request.

```bash
data='{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
curl -d "json=${data}" http://0.0.0.0:5000/route
```