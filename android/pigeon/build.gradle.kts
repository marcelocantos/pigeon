// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import org.gradle.internal.os.OperatingSystem

plugins {
    kotlin("jvm")
    `maven-publish`
}

group = "com.marcelocantos.pigeon"
version = "0.6.0"

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
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")
    testImplementation(kotlin("test"))
    testImplementation("tech.kwik:kwik:0.10.8")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.9.0")
}

// ----------------------------------------------------------------------------
// JNI native library build (T30).
//
// Compiles dist/pigeon.c + c/jni/pigeon_jni.c into libpigeon-jni for
// the host JVM platform. On macOS this produces .dylib, on Linux .so.
// Android NDK builds for arm64-v8a / x86_64 are deferred — the
// vendored libsodium AAR is a separate piece of work; production
// Android use is gated on shipping that AAR. See T30 commit notes.
//
// Resolution: the test JVM picks the lib up via java.library.path
// (set in tasks.test below); downstream consumers either bundle the
// .dylib/.so/.dll alongside their app or rely on
// `System.loadLibrary("pigeon-jni")` resolving from a well-known
// location.
// ----------------------------------------------------------------------------

val repoRoot = rootProject.projectDir.parentFile  // pigeon/
val distDir = file("$repoRoot/dist")
val jniSrcDir = file("$repoRoot/c/jni")
val vendorBuildDir = file("$repoRoot/c/vendor/build")
val sodiumIncludeDir = file("$vendorBuildDir/include")
val sodiumStaticLib = file("$vendorBuildDir/lib/libsodium.a")
val nativeOutDir = layout.buildDirectory.dir("native").get().asFile
val nativeLibName = "pigeon-jni"

val nativeLibFile: File by lazy {
    val ext = when {
        OperatingSystem.current().isMacOsX -> "dylib"
        OperatingSystem.current().isWindows -> "dll"
        else -> "so"
    }
    val prefix = if (OperatingSystem.current().isWindows) "" else "lib"
    File(nativeOutDir, "$prefix$nativeLibName.$ext")
}

val amalgamateNative = tasks.register<Exec>("amalgamateNative") {
    description = "Run `make amalgamate` so dist/pigeon.{c,h} exist before the JNI build."
    group = "build"
    workingDir = repoRoot
    commandLine = listOf("make", "amalgamate")
    inputs.dir(file("$repoRoot/c"))
    inputs.dir(file("$repoRoot/protocol"))
    outputs.file(file("$distDir/pigeon.c"))
    outputs.file(file("$distDir/pigeon.h"))
    isIgnoreExitValue = false
}

val buildVendoredSodium = tasks.register<Exec>("buildVendoredSodium") {
    description = "Build the vendored libsodium static lib (c/vendor/build/lib/libsodium.a)."
    group = "build"
    workingDir = repoRoot
    commandLine = listOf("bash", "c/vendor/build.sh", "libsodium")
    outputs.file(sodiumStaticLib)
    onlyIf { !sodiumStaticLib.exists() }
}

val compileNativeLibrary = tasks.register<Exec>("compileNativeLibrary") {
    description = "Build libpigeon-jni for the host platform (macOS .dylib / Linux .so)."
    group = "build"
    dependsOn(amalgamateNative)
    dependsOn(buildVendoredSodium)

    inputs.file("$distDir/pigeon.c")
    inputs.file("$distDir/pigeon.h")
    inputs.file("$jniSrcDir/pigeon_jni.c")
    inputs.file(sodiumStaticLib)
    outputs.file(nativeLibFile)

    doFirst {
        nativeOutDir.mkdirs()

        // JNI headers ship inside the JDK. Path layout differs slightly
        // per platform; we ask the configured Java toolchain.
        val javaHome = System.getProperty("java.home")
        val jniInclude = listOf(
            "$javaHome/include",
            "$javaHome/include/" + when {
                OperatingSystem.current().isMacOsX -> "darwin"
                OperatingSystem.current().isWindows -> "win32"
                else -> "linux"
            },
        )

        val cmd = mutableListOf(
            "clang",
            "-shared", "-fPIC", "-O2",
            "-DPIGEON_CRYPTO_LIBSODIUM",
            "-I$distDir",
            "-I$sodiumIncludeDir",
        )
        for (inc in jniInclude) cmd += "-I$inc"

        // Sources.
        cmd += "$distDir/pigeon.c"
        cmd += "$jniSrcDir/pigeon_jni.c"

        // Vendored libsodium static lib.
        cmd += sodiumStaticLib.absolutePath

        // Output.
        cmd += "-o"
        cmd += nativeLibFile.absolutePath

        commandLine = cmd
    }
}

tasks.named<Test>("test") {
    dependsOn(compileNativeLibrary)
    useJUnitPlatform()
    // Forward env vars to test JVM for live E2E tests.
    environment("PIGEON_TOKEN", System.getenv("PIGEON_TOKEN") ?: "")
    environment("PIGEON_RELAY_HOST", System.getenv("PIGEON_RELAY_HOST") ?: "")
    // Tell the JVM where to find libpigeon-jni.
    systemProperty("java.library.path", nativeOutDir.absolutePath)
    // The JNI bridge calls into pigeon_stream_send / _recv and
    // pigeon_datagram_send / _recv, each of which allocates a
    // ~1MB stack buffer (PIGEON_MAX_MSG). Default JVM thread stack
    // (~512K on macOS aarch64) is too small. Bump to 4M.
    jvmArgs("-Xss4m")
}

tasks.named<JavaCompile>("compileJava") {
    // Empty Kotlin-only module; nothing to do.
}

tasks.named("processResources") {
    // Bundle the host-built native library into the JAR so downstream
    // consumers (desktop JVM only) can extract it at startup. This is
    // a minimal scheme: production-ready libraries normally publish a
    // platform-classified jar, but the simple bundling proves the
    // distribution shape works.
    dependsOn(compileNativeLibrary)
}

// Copy the built native lib into the jar's resources/native/<os>/<arch>/.
val copyNativeLibIntoResources = tasks.register<Copy>("copyNativeLibIntoResources") {
    dependsOn(compileNativeLibrary)
    from(nativeLibFile)
    val arch = System.getProperty("os.arch") ?: "unknown"
    val os = when {
        OperatingSystem.current().isMacOsX -> "macos"
        OperatingSystem.current().isWindows -> "windows"
        else -> "linux"
    }
    into(layout.buildDirectory.dir("resources/main/native/$os/$arch"))
}

tasks.named("classes") {
    dependsOn(copyNativeLibIntoResources)
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
