# Seldon CLI

Seldon Core V2 can be managed via a CLI tool.

## Download Linux Binary

Download from a recent release from https://github.com/SeldonIO/seldon-core-v2/releases.

It is dynamically linked and will require and nix architetcure and glibc 2.25+.

```
mv seldon-linux-amd64 seldon
chmod u+x seldon
```

Add to your PATH.

## Local build (requires Go)

```bash
git clone https://github.com/SeldonIO/seldon-core-v2
cd seldon-core-v2/operator
make build-seldon
```

Add `<project-root>/operator/bin` to your PATH.




