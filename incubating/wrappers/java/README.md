# Java library for Seldon Core

A Java library to allow easy wrapping of Java models in Seldon Core.
See the docs on [how to
use](https://docs.seldon.io/projects/seldon-core/en/latest/java/README.html).

# Local Development Testing

Install to local repo

```
mvn install
```

Run your s2i build with your local MVN added as a volume with

```
--volume "$HOME/.m2":/root/.m2
```

An example s2i build would be:

```
s2i build --volume "$HOME/.m2":/root/.m2 model-template-app seldonio/seldon-core-s2i-java-build:0.1 myjavatest:0.1 --runtime-image seldonio/seldon-core-s2i-java-runtime:0.1
```
