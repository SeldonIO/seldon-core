# Test Executor with Logger

## REST

Run the following commands in different terminals.

Start dummy log sink.

```bash
make run_dummy_logsink
```

Start the executor locally.
```bash
make run_rest_executor
```

Start a dummy REST model locally.
```bash
make run_dummy_rest_model
```

Send a request
```bash
make curl_rest
```

The log sink should show the request payload as a Cloud Event:

```
{ path: '/',
  headers: 
   { host: 'localhost:2222',
     'user-agent': 'Go-http-client/1.1',
     'content-length': '32',
     'ce-id': 'a25fcefe-238f-4a3e-972a-fe64419ca74f',
     'ce-source': 'http://localhost:8000/',
     'ce-specversion': '0.2',
     'ce-time': '2019-12-24T17:55:29.709146122Z',
     'ce-type': 'io.seldon.serving.inference.request',
     'content-type': 'application/json',
     'model-id': 'classifier',
     'accept-encoding': 'gzip' },
  method: 'POST',
  body: '{"data":{"ndarray":[[1.0,2.0]]}}',
  cookies: undefined,
  fresh: false,
  hostname: 'localhost',
  ip: '::ffff:172.17.0.1',
  ips: [],
  protocol: 'http',
  query: {},
  subdomains: [],
  xhr: false,
  os: { hostname: '9865dd6ba322' } }

```


## gRPC

Run the following commands in different terminals.

Start dummy log sink.

```bash
make run_dummy_logsink
```

Start the executor locally.
```bash
make run_grpc_executor
```

Start a dummy REST model locally.
```bash
make run_dummy_grpc_model
```

Send a request
```bash
make grpc_test
```

The log sink should show the request payload as a Cloud Event:

```
{ path: '/',
  headers: 
   { host: 'localhost:2222',
     'user-agent': 'Go-http-client/1.1',
     'content-length': '42',
     'ce-id': '0a032dc1-9882-4188-9453-f7c5be386345',
     'ce-source': 'http://localhost:8000/',
     'ce-specversion': '0.2',
     'ce-time': '2019-12-24T18:00:04.082966884Z',
     'ce-type': 'io.seldon.serving.inference.request',
     'model-id': 'classifier',
     'accept-encoding': 'gzip' },
  method: 'POST',
  body: '"GhwaGgoYMhYKCREAAAAAAADwPwoJEQAAAAAAAABA"',
  cookies: undefined,
  fresh: false,
  hostname: 'localhost',
  ip: '::ffff:172.17.0.1',
  ips: [],
  protocol: 'http',
  query: {},
  subdomains: [],
  xhr: false,
  os: { hostname: 'fc40e5bd5581' } }
```

