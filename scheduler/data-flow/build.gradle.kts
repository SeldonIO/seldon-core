import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    kotlin("jvm") version "1.6.10"
    application
}

group = "io.seldon"
version = "1.0-SNAPSHOT"

repositories {
    mavenLocal()
    mavenCentral()
    maven {
        url = uri("https://packages.confluent.io/maven/")
    }
}

dependencies {
    implementation("org.jetbrains.kotlin:kotlin-gradle-plugin:1.6.10")
    implementation("com.natpryce:konfig:1.6.10.0")
    implementation("io.klogging:klogging-jvm:0.4.4")

    // Kafka
    implementation("org.apache.kafka:kafka-streams:3.1.0")
    implementation("io.confluent:kafka-streams-protobuf-serde:7.0.1")
    implementation("io.klogging:slf4j-klogging:0.2.5")

    // gRPC
    implementation("io.grpc:grpc-kotlin-stub:1.2.1")
    implementation("io.grpc:grpc-stub:1.45.0")
    implementation("io.grpc:grpc-protobuf:1.45.0")
    runtimeOnly("io.grpc:grpc-netty-shaded:1.44.1")
    implementation("com.google.protobuf:protobuf-java:3.19.4")
    implementation("com.google.protobuf:protobuf-kotlin:3.19.4")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.6.0")

    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter-params:5.8.2")
    testImplementation("io.strikt:strikt-core:0.34.1")
}

sourceSets {
    main {
        java {
            srcDirs("src/main/kotlin")
        }
    }
}

tasks.test {
    useJUnitPlatform()
}

java {
    toolchain {
        languageVersion.set(JavaLanguageVersion.of(17))
    }
}

tasks.withType<KotlinCompile> {
    kotlinOptions {
        jvmTarget = "17"
        freeCompilerArgs += "-Xopt-in=kotlin.RequiresOptIn"
    }
}

val dataflowMainClass = "io.seldon.dataflow.Main"

tasks.withType<Jar> {
    manifest {
        attributes["Main-Class"] = dataflowMainClass
    }

    from(
        configurations.runtimeClasspath
            .get()
            .map { if (it is Directory) it else zipTree(it) }
    )

    duplicatesStrategy = DuplicatesStrategy.EXCLUDE
}

application {
    mainClass.set(dataflowMainClass)
}