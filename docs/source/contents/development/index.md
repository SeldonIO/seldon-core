# Development

## Development Requirements

 * Go 1.17+ with glibc development libraries installed.
    * e.g. for Ubuntu systems: `sudo apt-get install libc6-dev`
 * Java JDK 17+ and Kotlin 1.6.10+
 * Kubebuilder V2
 * Docker and docker-compose
 * Helm
 * Kustomize
 * Ansible

## Testing resources

 * Kind
 * k6
 * unix utils: jq, curl, grpcurl

## Release Process

[Releases are carried out via Github.](./release/index.md)

```{toctree}
:maxdepth: 1
:hidden:

release/index.md
licenses.md
```
