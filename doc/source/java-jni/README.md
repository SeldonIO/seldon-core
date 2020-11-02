# Packaging a Java model for Seldon Core using s2i

In this guide, we illustrate the steps needed to wrap your own Java model in a
docker image ready for deployment with Seldon Core using [source-to-image app
s2i](https://github.com/openshift/source-to-image) and the JNI wrapper.

The JNI wrapper leverages the [Python inference server](../python) to
communicate through JNI with your Java model.
The Python wrapper usually receives updates much faster than the [native Java
one](../java/README.md).
Therefore it allows you to have faster access to the latest Seldon Core
features.

If you are not familiar with s2i you can read [general instructions on using
s2i](../wrappers/s2i.md) and then follow the steps below.

## Step 1 - Install s2i

[Download and install s2i](https://github.com/openshift/source-to-image#installation)

 * Prerequisites for using s2i are:
   * Docker
   * Git (if building from a remote git repo)

To check everything is working you can run

```bash
s2i usage seldonio/s2i-java-jni-build:0.3.0
```

## Step 2 - Create your source code

To use our s2i builder image to package your Java model you will need:

 * A Maven project that depends on `io.seldon.wrapper` library (>= `0.3.0`)
 * A class that implements `io.seldon.wrapper.api.SeldonPredictionService` for the type of component you are creating
 * .s2i/environment - model definitions used by the s2i builder to correctly wrap your model

We will go into detail for each of these steps:

### Maven Project
Create a Spring Boot Maven project and include the dependency:

```XML
<dependency>
	<groupId>io.seldon.wrapper</groupId>
	<artifactId>seldon-core-wrapper</artifactId>
	<version>0.3.0</version>
</dependency>
```

A full example can be found at `incubating/wrappers/s2i/java-jni/test/model-template-app/pom.xml`.

### Prediction Class
To handle requests to your model or other component you need to implement one
or more of the methods in `io.seldon.wrapper.api.SeldonPredictionService`, in
particular:

```java
default public SeldonMessage predict(SeldonMessage request);
default public SeldonMessage route(SeldonMessage request);
default public SeldonMessage sendFeedback(Feedback request);
default public SeldonMessage transformInput(SeldonMessage request);
default public SeldonMessage transformOutput(SeldonMessage request);
default public SeldonMessage aggregate(SeldonMessageList request);
```

There is a full H2O example in
`examples/models/h2o_mojo/src/main/java/io/seldon/example/h2o/model`, whose
implementation is shown below:

```java
public class H2OModelHandler implements SeldonPredictionService {
	private static Logger logger = LoggerFactory.getLogger(H2OModelHandler.class.getName());
	EasyPredictModelWrapper model;

	public H2OModelHandler() throws IOException {
		MojoReaderBackend reader =
                MojoReaderBackendFactory.createReaderBackend(
                  getClass().getClassLoader().getResourceAsStream(
                     "model.zip"),
                      MojoReaderBackendFactory.CachingStrategy.MEMORY);
		MojoModel modelMojo = ModelMojoReader.readFrom(reader);
		model = new EasyPredictModelWrapper(modelMojo);
		logger.info("Loaded model");
	}

	@Override
	public SeldonMessage predict(SeldonMessage payload) {
		List<RowData> rows = H2OUtils.convertSeldonMessage(payload.getData());
		List<AbstractPrediction> predictions = new ArrayList<>();
		for(RowData row : rows)
		{
			try
			{
				BinomialModelPrediction p = model.predictBinomial(row);
				predictions.add(p);
			} catch (PredictException e) {
				logger.info("Error in prediction ",e);
			}
		}
        DefaultData res = H2OUtils.convertH2OPrediction(predictions, payload.getData());

		return SeldonMessage.newBuilder().setData(res).build();
	}

}

```

The above code:

  * Loads a model from the local resources folder on startup
  * Converts the proto buffer message into H2O RowData using provided utility classes.
  * Runs a BionomialModel prediction and converts the result back into a `SeldonMessage` for return

#### H2O Helper Classes

We provide H2O utility class `io.seldon.wrapper.utils.H2OUtils` in
seldon-core-wrapper to convert to and from the seldon-core proto buffer message
types.

#### DL4J Helper Classes

We provide a DL4J utility class `io.seldon.wrapper.utils.DL4JUtils` in
seldon-core-wrapper to convert to and from the seldon-core proto buffer message
types.

### .s2i/environment

Define the core parameters needed by our Java S2I images to wrap your model. 
An example is:

```bash
SERVICE_TYPE=MODEL
JAVA_IMPORT_PATH=io.seldon.example.model.ExampleModelHandler
```

These values can also be provided or overridden on the command line when building the image.

## Step 3 - Build your image

Use `s2i build` to create your Docker image from source code.
You will need Docker installed on the machine and optionally git if your source
code is in a public git repo.

Using s2i you can build directly from a git repo or from a local source folder.
See the [s2i
docs](https://github.com/openshift/source-to-image/blob/master/docs/cli.md#s2i-build)
for further details.
The general format is:

```bash
s2i build \
  <git-repo | src-folder> \
  seldonio/s2i-java-jni-build:0.3.0 \
  --runtime-image seldonio/s2i-java-jni-runtime:0.3.0 \
  <my-image-name> 
```

An example invocation using the test template model inside `seldon-core`:

```bash
s2i build \
  https://github.com/seldonio/seldon-core.git \
  --context-dir=incubating/wrappers/s2i/java-jni/test/model-template-app \
  seldonio/s2i-java-jni-build:0.3.0 \
  --runtime-image seldonio/s2i-java-jni-runtime:0.3.0 \
  jni-model-template:0.1.0
```

The above s2i build invocation:

 * uses the `seldon-core` [GitHub repo](https://github.com/seldonio/seldon-core)
   and the directory `incubating/wrappers/s2i/java-jni/test/model-template-app`
   inside that repo.
 * uses the builder image `seldonio/s2i-java-jni-build`
 * uses the runtime image `seldonio/s2i-java-jni-runtime`
 * creates a docker image `jni-template-model`


For building from a local source folder, an example where we clone the seldon-core repo:

```bash
git clone https://github.com/seldonio/seldon-core.git
cd seldon-core
s2i build \
  incubating/wrappers/s2i/java-jni/test/model-template-app \
  seldonio/s2i-java-jni-build:0.3.0 \
  --runtime-image seldonio/s2i-java-jni-runtime:0.3.0 \
  jni-model-template:0.1.0
```

For more help see:

```bash
s2i usage seldonio/s2i-java-jni-build:0.3.0
s2i build --help
```

## Reference

## Environment Variables
The required environment variables understood by the builder image are
explained below.
You can provide them in the `.s2i/environment` file or on the `s2i build`
command line.

### JAVA_IMPORT_PATH

Import path for your Java model implementation.
For instance, in the example above, this would be
`io.seldon.example.model.ExampleModelHandler`.


### SERVICE_TYPE

The service type being created.
Available options are:

 * MODEL
 * ROUTER
 * TRANSFORMER
 * COMBINER


## Creating different service types

### MODEL

 * [A minimal skeleton for model source code](https://github.com/SeldonIO/seldon-core/tree/master/incubating/wrappers/s2i/java/test/model-template-app)
 * [Example H2O MOJO](../examples/h2o_mojo.html)
