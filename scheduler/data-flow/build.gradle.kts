import com.github.jengelman.gradle.plugins.shadow.tasks.ShadowJar
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile


plugins {
    id("com.github.hierynomus.license-report") version "0.16.1"
    id("com.github.johnrengelman.shadow") version "8.1.1"
    kotlin("jvm") version "2.1.0" // the kotlin version
    kotlin("plugin.serialization") version "2.1.0" // kotlinx serialization plugin
    id("org.jlleitschuh.gradle.ktlint") version "13.1.0"
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
    implementation("io.klogging:klogging-jvm:0.11.1")
    implementation("io.klogging:slf4j-klogging:0.11.1")

    // Kafka
    implementation("org.apache.kafka:kafka-streams:8.0.0-ccs")
    testImplementation("org.apache.kafka:kafka-streams-test-utils:8.0.0-ccs")
    // https://mvnrepository.com/artifact/io.confluent/kafka-streams-protobuf-serde
    implementation("io.confluent:kafka-streams-protobuf-serde:8.0.0")

    // gRPC
    implementation("io.grpc:grpc-kotlin-stub:1.4.3")
    implementation("io.grpc:grpc-stub:1.73.0")
    implementation("io.grpc:grpc-protobuf:1.73.0")
    runtimeOnly("io.grpc:grpc-netty-shaded:1.73.0")
    implementation("com.google.protobuf:protobuf-java") {
        version {
            strictly("[4.29.3,)")
            prefer("4.29.3")
        }
    }
    implementation("com.google.protobuf:protobuf-kotlin") {
        version {
            strictly("[4.29.3,)")
            prefer("4.29.3")
        }
    }
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.10.2")
    implementation("com.michael-bull.kotlin-retry:kotlin-retry:2.0.1")

    // k8s
    implementation("io.kubernetes:client-java:24.0.0")

    // HTTP server for health probes
    implementation("io.ktor:ktor-server-core:3.3.0")
    implementation("io.ktor:ktor-server-netty:3.3.0")
    implementation("io.ktor:ktor-server-content-negotiation:3.3.0")
    implementation("io.ktor:ktor-serialization-kotlinx-json:3.3.0")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.7.3")
    implementation("com.charleskorn.kaml:kaml:0.61.0")

    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter-params:5.10.3")
    testImplementation("io.strikt:strikt-core:0.34.1")
    testImplementation("org.assertj:assertj-core:3.25.1")
}

val generatedSourceDir = layout.buildDirectory.dir("generated/source/buildinfo")

sourceSets {
    main {
        java {
            srcDirs("src/main/kotlin")
            srcDir(generatedSourceDir)
        }
    }
}

val generateBuildInfo by tasks.registering {
    val outputDir = generatedSourceDir.get().asFile
    outputs.dir(outputDir)

    doLast {
        val buildInfoFile = File(outputDir, "io/seldon/dataflow/BuildInfo.kt")
        buildInfoFile.parentFile.mkdirs()
        val releaseTag = project.findProperty("release_tag")?.toString() ?: "unknown"

        buildInfoFile.writeText(
            """
            package io.seldon.dataflow

            object BuildInfo {
                const val VERSION = "$releaseTag"
            }
            """.trimIndent(),
        )
    }
}

tasks.test {
    useJUnitPlatform()
    testLogging {
        events("PASSED", "SKIPPED", "FAILED")
    }
}

java {
    toolchain {
        languageVersion.set(JavaLanguageVersion.of(17))
    }
}

tasks.withType<KotlinCompile> {
    dependsOn(generateBuildInfo)
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

tasks.withType<org.jlleitschuh.gradle.ktlint.tasks.KtLintCheckTask> {
    dependsOn(generateBuildInfo)
}

tasks.withType<org.jlleitschuh.gradle.ktlint.tasks.KtLintFormatTask> {
    dependsOn(generateBuildInfo)
}

ktlint {
    verbose = true
    debug = true
    // Ignore generated code from proto
    filter {
        exclude { element -> element.file.path.contains("apis/mlops") }
        exclude { element -> element.file.path.contains("generated/source/buildinfo") }
    }
}
