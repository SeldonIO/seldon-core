# Kubernetes Client

We use the official Java Kubernetes client. To insure for correct use make sure all 3 java projects are updated at same time to new versions in pom.xml,

```
<dependency>
	<groupId>io.kubernetes</groupId>
	<artifactId>client-java</artifactId>
	<version>3.0.0</version>
	<scope>compile</scope>
</dependency>
```

Also ensure the part of the pom.xml that compiles the proto buffers has `excludes` for any Kubernetes classes that would be found in the Kubernetes Java client:

```
<plugin>
	<groupId>org.xolstice.maven.plugins</groupId>
	<artifactId>protobuf-maven-plugin</artifactId>
	<version>0.5.0</version>
	<configuration>
		<protocArtifact>com.google.protobuf:protoc:3.1.0:exe:${os.detected.classifier}</protocArtifact>
		<pluginId>grpc-java</pluginId>
		<pluginArtifact>io.grpc:protoc-gen-grpc-java:${grpc.version}:exe:${os.detected.classifier}</pluginArtifact>
		<clearOutputDirectory>false</clearOutputDirectory>
		<excludes>
			<exclude>k8s.io/**/*.proto</exclude>
			<exclude>**/v1.proto</exclude>
		</excludes>
	</configuration>
	<executions>
		<execution>
			<goals>
				<goal>compile</goal>
				<goal>compile-custom</goal>
			</goals>
		</execution>
	</executions>
</plugin>
```