plugins {
    kotlin("jvm")
    `maven-publish`
}

group = "com.marcelocantos.pigeon"
version = "0.5.0"

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
    testImplementation(kotlin("test"))
    testImplementation("tech.kwik:kwik:0.10.8")
}

tasks.test {
    useJUnitPlatform()
    // Forward env vars to test JVM for live E2E tests.
    environment("PIGEON_TOKEN", System.getenv("PIGEON_TOKEN") ?: "")
    environment("PIGEON_RELAY_HOST", System.getenv("PIGEON_RELAY_HOST") ?: "")
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            from(components["java"])
            groupId = "com.marcelocantos.pigeon"
            artifactId = "pigeon"
        }
    }
}
