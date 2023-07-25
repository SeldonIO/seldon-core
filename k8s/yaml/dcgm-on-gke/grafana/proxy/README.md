## Building the Inverse Proxy Agent

Note that the inverse proxy agent relies on a proxy server hosted on AppEngine. This should be sufficient for most users. For more information about setting up your own inverse proxy server on AppEngine see https://github.com/google/inverting-proxy. 

The inverse proxy agent deployed to the GKE cluster uses a Docker Hub image hosted at sukha/inverse-proxy-for-grafana.

You can build and host this same image yourself:
```
gcloud artifacts repositories create REPOSITORY_NAME \
  --repository-format=docker \
  --location=REGION

gcloud auth configure-docker REGION-docker.pkg.dev

cd grafana/proxy
docker build -f Dockerfile -t IMAGE_NAME .
docker push REGION-docker.pkg.dev/PROJECT_ID/REPOSITORY_NAME/IMAGE_NAME
```

Here you need to fill the appropriate values for REPOSITORY_NAME, REGION, PROJECT_ID, and IMAGE_NAME.


Then modify grafana.yml to select the inverse proxy agent you just created:
```
apiVersion: v1
kind: Pod
metadata:
  name: inverse-proxy
  namespace: gpu-monitoring-system
spec:
  containers:
  - name: inverse-proxy
    image: REGION-docker.pkg.dev/PROJECT_ID/REPOSITORY_NAME/IMAGE_NAME
    ports:
    - containerPort: 80
```
