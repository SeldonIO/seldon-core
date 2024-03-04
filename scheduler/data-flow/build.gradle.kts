import org.jetbrains.kotlin.gradle.tasks.KotlinCompile
import com.github.jengelman.gradle.plugins.shadow.tasks.ShadowJar


plugins {
    id("com.github.hierynomus.license-report") version "0.16.1"
    id("com.github.johnrengelman.shadow") version "8.1.1"
    kotlin("jvm") version "1.8.20"

    java
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
    implementation("com.natpryce:konfig:1.6.10.0")
    implementation("io.klogging:klogging-jvm:0.5.11")
    implementation("io.klogging:slf4j-klogging:0.5.11")

    // Kafka
    implementation("org.apache.kafka:kafka-streams:7.6.0-ccs")
    testImplementation("org.apache.kafka:kafka-streams-test-utils:7.6.0-ccs")

    // gRPC
    implementation("io.grpc:grpc-kotlin-stub:1.4.1")
    implementation("io.grpc:grpc-stub:1.61.1")
    implementation("io.grpc:grpc-protobuf:1.61.1")
    runtimeOnly("io.grpc:grpc-netty-shaded:1.62.2")
    implementation("com.google.protobuf:protobuf-java:3.25.2")
    implementation("com.google.protobuf:protobuf-kotlin:3.25.2")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.7.3")
    implementation("com.michael-bull.kotlin-retry:kotlin-retry:1.0.9")

    // k8s
    implementation("io.kubernetes:client-java:19.0.0")

    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter-params:5.10.2")
    testImplementation("io.strikt:strikt-core:0.34.1")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
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
        freeCompilerArgs += "-opt-in=kotlin.RequiresOptIn"
    }
}

val dataflowMainClass = "io.seldon.dataflow.Main"

application {
    mainClass.set(dataflowMainClass)
}

tasks.named<ShadowJar>("shadowJar") {
    mergeServiceFiles()
}

downloadLicenses {
    includeProjectDependencies = true
    dependencyConfiguration = "compileClasspath"
}
