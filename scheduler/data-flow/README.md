# Inference Graph Data Flow - Kafka Streams

## About

The data-flow transformer is built on top of [Kafka Streams](https://docs.confluent.io/platform/current/streams/index.html).
It implements the operations required to propagate information throughout an inference pipeline in Seldon Core v2.
These operations include multi-way joins, or fan-ins, and conversion of inference v2 response types to request types.

The core idea of asynchronous data flows is that they offer:
* buffering during periods of bursty traffic, e.g. for black-box explainers
* dynamic reconfiguration of pipelines to incorporate advanced monitoring components, e.g. drift detectors
* replay of messages
* auditability

## Developer setup

You will need:
* JDK 17+
* Kotlin 1.6.10+

### JDK

You can install JDK versions using:
* [sdkman](https://sdkman.io/)
* [asdf](https://asdf-vm.com/guide/getting-started.html)

To install with `asdf`, run:

```bash
asdf plugin-add java
asdf list-all java | less                   # See available options
ASDF_JAVA_VERSION=graalvm-22.0.0.2+java17   # Or your chosen JDK, `zulu-17.32.13` should also work
asdf install java $ASDF_JAVA_VERSION
asdf global java $ASDF_JAVA_VERSION
```

GraalVM offers strong performance through intelligent inlining and optimisations.
Zulu is stable, well supported too, and has performant GC behaviour.
Other options, such as `Coretto`, `AdoptOpenJDK`, and plain old Oracle `OpenJDK` also exist.

### Kotlin

You can install a Kotlin SDK using `asdf` or `sdkman`.
Alternatively, you can simply let Gradle do the work... Read on.

### Gradle

There is a Gradle wrapper included in this repository: `./gradlew`.
You can install the necessary build and compilation dependencies by running the `build` target:

```bash
$ ./gradlew build

BUILD SUCCESSFUL in 536ms
6 actionable tasks: 6 up-to-date
```

<details>
<summary>Unsupported class file major version</summary>

If you see an error like the below, your **Gradle** version is not high enough.

```
_BuildScript_' Unsupported class file major version 61
```

Check for a new enough version for your Java and Kotlin SDK versions.
Update the version of `distributionUrl` in `./gradle/wrapper/gradle-wrapper.properties`.

If you run `./gradlew build` again, it should download new dependencies and progress.

If you are in JetBrains IDEA, you will need to close and re-open the project, as it gets confused.
</details>

<details>
<summary>Unknown JVM target version</summary>

If you see an error like the below, your **Kotlin** version is not high enough.

```
Task :compileKotlin FAILED
e: Unknown JVM target version: 17
```

Check for a newer version and update the version in `build.gradle.kts`:

```
plugins {
  ...
  kotlin("jvm") version "X.X.X"
}
```
</details>

### gRPC protocol buffers

Rather than the effort of setting up a Gradle plugin, it is straightforward to use `protoc` directly.
The steps for Kotlin are [documented here](https://github.com/grpc/grpc-kotlin/blob/master/compiler/README.md#manual-protoc-usage).

You will need both the Java and Kotlin extensions for `protoc`, available at:
* [https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-kotlin/](https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-kotlin/)
* [https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/](https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/)

If running manually, the command will look like:

```bash
cd $(git rev-parse --show-toplevel)/apis/mlops/chainer
protoc \
  --proto_path=. \
  --plugin=protoc-gen-grpc-java=<path to protoc-gen-grpc-java> \
  --java_out=./kotlin \
  --grpc-java_out=./kotlin \
  --plugin=protoc-gen-grpckt=<path to protoc-gen-grpc-kotlin.sh> \
  --kotlin_out=./kotlin \
  --grpckt_out=./kotlin \
  chainer.proto
```