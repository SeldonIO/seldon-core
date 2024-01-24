import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
    id("com.github.hierynomus.license-report") version "0.16.1"
    kotlin("jvm") version "1.6.21"
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
    implementation("io.klogging:klogging-jvm:0.4.4")

    // Kafka
    implementation("org.apache.kafka:kafka-streams:3.4.0")
    implementation("io.klogging:slf4j-klogging:0.2.5")
    testImplementation("org.apache.kafka:kafka-streams-test-utils:3.5.0")

    // gRPC
    implementation("io.grpc:grpc-kotlin-stub:1.2.1")
    implementation("io.grpc:grpc-stub:1.57.2")
    implementation("io.grpc:grpc-protobuf:1.57.2")
    runtimeOnly("io.grpc:grpc-netty-shaded:1.44.1")
    implementation("com.google.protobuf:protobuf-java:3.21.7")
    implementation("com.google.protobuf:protobuf-kotlin:3.21.7")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.6.0")
    implementation("com.michael-bull.kotlin-retry:kotlin-retry:1.0.9")

    // k8s
    implementation("io.kubernetes:client-java:19.0.0")

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
        freeCompilerArgs += "-opt-in=kotlin.RequiresOptIn"
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
    ) { exclude("META-INF/*.RSA", "META-INF/*.SF", "META-INF/*.DSA")}

    duplicatesStrategy = DuplicatesStrategy.EXCLUDE

}

application {
    mainClass.set(dataflowMainClass)
}


downloadLicenses {
    includeProjectDependencies = true
    dependencyConfiguration = "compileClasspath"
}
