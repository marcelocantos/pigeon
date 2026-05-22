// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import org.gradle.internal.os.OperatingSystem

plugins {
    id("com.android.library")
    kotlin("android")
    `maven-publish`
}

group = "com.marcelocantos.pigeon"
version = "0.6.0"

// ---------------------------------------------------------------------------
// Layout: where pigeon's C sources, amalgamated dist, and vendored libsodium
// live relative to this Gradle module.
// ---------------------------------------------------------------------------
val repoRoot = rootProject.projectDir.parentFile          // pigeon/
val distDir = file("$repoRoot/dist")                       // amalgamated pigeon.c/h
val jniSrcDir = file("$repoRoot/c/jni")                    // pigeon_jni.c (T36 may extend)
val vendorBuildDir = file("$repoRoot/c/vendor/build")      // host (desktop) libsodium
val sodiumStaticLib = file("$vendorBuildDir/lib/libsodium.a")
val sodiumIncludeDir = file("$vendorBuildDir/include")
val androidSodiumRoot = file("$repoRoot/c/vendor/build-android")

// AGP-shipped ABIs we support. arm64-v8a is the production target;
// x86_64 covers the standard emulator. Add armeabi-v7a / x86 only on
// explicit demand — 32-bit Android phones are well out of the floor.
val supportedAbis = listOf("arm64-v8a", "x86_64")

android {
    namespace = "com.marcelocantos.pigeon"
    // compileSdk 34 is the floor that gives us JDK 17 toolchain support
    // without forcing a brand-new toolchain on every contributor. The
    // SDK requirement note in CLAUDE.md ("Android API 33+ for X25519")
    // is enforced via minSdk below.
    compileSdk = 34

    // Pin to an installed NDK. Override via local.properties (ndk.dir)
    // or ANDROID_NDK_HOME if a contributor has a different revision.
    // 27.x and 29.x both work; we test on 29.0.14206865.
    ndkVersion = (
        project.findProperty("pigeon.ndkVersion") as String?
    ) ?: "29.0.14206865"

    defaultConfig {
        minSdk = 33  // X25519 lands in API 33.
        ndk {
            abiFilters += supportedAbis
        }
        externalNativeBuild {
            cmake {
                arguments += listOf(
                    "-DANDROID_STL=none",
                    "-DPIGEON_REPO_ROOT=${repoRoot.absolutePath}",
                )
            }
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    sourceSets {
        // The pre-existing module uses src/main/kotlin and src/test/kotlin
        // (the kotlin("jvm") convention). AGP also accepts those, but we
        // make it explicit so a fresh checkout doesn't silently miss
        // files.
        named("main") {
            java.setSrcDirs(listOf("src/main/kotlin"))
        }
        named("test") {
            java.setSrcDirs(listOf("src/test/kotlin"))
        }
    }

    externalNativeBuild {
        cmake {
            path = file("src/main/cpp/CMakeLists.txt")
            version = "3.22.1"
        }
    }

    packaging {
        // The two ABIs already deduplicate by directory; nothing to
        // pickFirst here. If a downstream consumer brings its own
        // libsodium, the static-link inside libpigeon-jni.so insulates
        // them from symbol collisions.
    }

    // AGP local JVM unit tests run via :testDebugUnitTest (and the
    // aggregate :test). They use the desktop JVM, so they need
    // libpigeon-jni.dylib/.so on java.library.path — see the
    // configuration of the unit-test task below.
    testOptions {
        unitTests {
            isIncludeAndroidResources = false
            all {
                it.useJUnitPlatform()
                // The JNI bridge calls into pigeon_stream_send / _recv
                // and pigeon_datagram_send / _recv, each allocating a
                // ~1MB stack buffer (PIGEON_MAX_MSG). Default JVM
                // thread stack (~512K on macOS aarch64) is too small.
                it.jvmArgs("-Xss4m")
            }
        }
    }

    publishing {
        singleVariant("release") {
            withSourcesJar()
        }
    }
}

dependencies {
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")
    testImplementation(kotlin("test"))
    testImplementation("tech.kwik:kwik:0.10.8")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.9.0")
}

// ---------------------------------------------------------------------------
// Desktop JVM JNI build (.dylib/.so/.dll) — used by local unit tests.
//
// This mirrors the pre-T35 build: it compiles dist/pigeon.c + the JNI
// shim against the host-platform libsodium static lib produced by
// c/vendor/build.sh. AGP's externalNativeBuild path covers the Android
// ABIs separately; both paths coexist so :test runs locally and
// :assembleRelease produces the AAR with .so files inside.
// ---------------------------------------------------------------------------

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
    // Use narrowly-scoped inputs so we don't accidentally claim
    // ownership of c/vendor/ (which the libsodium build tasks own).
    // Gradle's input snapshotting would otherwise flag a phantom
    // dependency between amalgamateNative and buildVendoredSodium.
    inputs.dir(file("$repoRoot/c/src"))
    inputs.dir(file("$repoRoot/c/include"))
    inputs.file(file("$repoRoot/c/amalgamate.sh"))
    inputs.dir(file("$repoRoot/protocol"))
    outputs.file(file("$distDir/pigeon.c"))
    outputs.file(file("$distDir/pigeon.h"))
    isIgnoreExitValue = false
}

val buildVendoredSodium = tasks.register<Exec>("buildVendoredSodium") {
    description = "Build the host-platform vendored libsodium static lib (desktop JVM tests)."
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
        cmd += "$distDir/pigeon.c"
        cmd += "$jniSrcDir/pigeon_jni.c"
        cmd += sodiumStaticLib.absolutePath
        cmd += "-o"
        cmd += nativeLibFile.absolutePath
        commandLine = cmd
    }
}

// ---------------------------------------------------------------------------
// Android NDK build — cross-compile libsodium per ABI.
//
// AGP's externalNativeBuild { cmake { } } takes care of compiling
// pigeon.c + pigeon_jni.c per ABI into libpigeon-jni.so, then bundles
// those into the AAR under jni/<abi>/. But the CMakeLists.txt in
// src/main/cpp/ depends on the vendored libsodium static lib being
// present at c/vendor/build-android/<abi>/lib/libsodium.a. This task
// ensures that prerequisite, calling out to c/vendor/build-android.sh.
//
// One task per ABI so Gradle's build cache tracks per-ABI outputs and
// we get a clean serial log instead of an interleaved mess. Both
// invocations share the libsodium source tree, so they are NOT safe to
// parallelise — the script is sequential by design.
// ---------------------------------------------------------------------------

val buildAndroidSodiumTasks = supportedAbis.map { abi ->
    tasks.register<Exec>("buildAndroidSodium-$abi") {
        description = "Cross-compile vendored libsodium for Android ABI $abi."
        group = "build"
        workingDir = repoRoot
        commandLine = listOf("bash", "c/vendor/build-android.sh", abi)
        outputs.file(file("$androidSodiumRoot/$abi/lib/libsodium.a"))
        outputs.dir(file("$androidSodiumRoot/$abi/include"))
        onlyIf {
            !file("$androidSodiumRoot/$abi/lib/libsodium.a").exists()
        }
    }
}

val buildAndroidSodium = tasks.register("buildAndroidSodium") {
    description = "Cross-compile vendored libsodium for all supported Android ABIs."
    group = "build"
    dependsOn(buildAndroidSodiumTasks)
}

// AGP generates ABI-specific CMake configure tasks lazily; wire each
// of them to depend on its corresponding libsodium build and on the
// amalgamated pigeon source. The task names follow AGP's convention:
//   configureCMake<Variant>[Debug|Release]<Abi>
// We match by simple-name prefix to avoid coupling to the exact
// internal task class.
afterEvaluate {
    tasks.matching {
        val n = it.name
        n.startsWith("configureCMake") ||
            n.startsWith("buildCMake")  // covers buildCMakeDebug / buildCMakeRelWithDebInfo
    }.configureEach {
        dependsOn(amalgamateNative)
        dependsOn(buildAndroidSodium)
    }

    // Unit tests need the desktop .dylib on java.library.path.
    tasks.withType<Test>().configureEach {
        dependsOn(compileNativeLibrary)
        systemProperty("java.library.path", nativeOutDir.absolutePath)
        // Forward env vars for live E2E tests.
        environment("PIGEON_TOKEN", System.getenv("PIGEON_TOKEN") ?: "")
        environment("PIGEON_RELAY_HOST", System.getenv("PIGEON_RELAY_HOST") ?: "")
    }
}

// ---------------------------------------------------------------------------
// Publishing — AAR via maven-publish, consumed by JitPack.
// ---------------------------------------------------------------------------
publishing {
    publications {
        register<MavenPublication>("release") {
            // AGP creates the "release" software component after evaluation.
            afterEvaluate {
                from(components["release"])
            }
            groupId = "com.marcelocantos.pigeon"
            artifactId = "pigeon"
        }
    }
}
