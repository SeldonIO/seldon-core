# Seldon CLI

Seldon Core V2 can be managed via a CLI tool.

## Download Linux Binary

Download from a recent release from https://github.com/SeldonIO/seldon-core/releases.

It is dynamically linked and will require and *nix architecture and glibc 2.25+.

```
mv seldon-linux-amd64 seldon
chmod u+x seldon
```

Add to your PATH.

## Local build (requires Go)

```bash
git clone https://github.com/SeldonIO/seldon-core --branch=v2
cd seldon-core/operator
make build-seldon
```

Add `<project-root>/operator/bin` to your PATH.

## Local macOS ARM build (requires Go and librdkafka)

```bash
# install dependencies
brew install go librdkafka
```

```bash
git clone https://github.com/SeldonIO/seldon-core --branch=v2
cd seldon-core/operator
make build-seldon-arm
```

Add `<project-root>/operator/bin` to your PATH.




