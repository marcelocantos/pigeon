plugins {
    kotlin("jvm")
    `maven-publish`
}

group = "com.marcelocantos.tern"
version = "0.3.0"

java {
    sourceCompatibility = JavaVersion.VERSION_21
    targetCompatibility = JavaVersion.VERSION_21
}

kotlin {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_21)
    }
}

dependencies {
    implementation(project(":terncrypto"))
    testImplementation(kotlin("test"))
    testImplementation("tech.kwik:kwik:0.10.8")
}

tasks.test {
    useJUnitPlatform()
    // Forward env vars to test JVM for live E2E tests.
    environment("TERN_TOKEN", System.getenv("TERN_TOKEN") ?: "")
    environment("TERN_RELAY_HOST", System.getenv("TERN_RELAY_HOST") ?: "")
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            from(components["java"])
            groupId = "com.marcelocantos.tern"
            artifactId = "ternrelay"
        }
    }
}
