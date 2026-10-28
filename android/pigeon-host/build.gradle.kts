// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Minimal com.android.application host target for the :pigeon library's
// device instrumentation suite (🎯T66.6). A com.android.library module
// never produces a debug APK — only an AAR plus a self-contained
// androidTest APK — so scripts/android-device-e2e.sh had nothing
// installable to use as the instrumentation target. This module depends
// on :pigeon and installs as a real app package; the androidTest sources
// that used to self-instrument inside :pigeon now self-instrument here,
// exercising the library's public Kotlin API exactly as before.
plugins {
    id("com.android.application")
    kotlin("android")
}

android {
    namespace = "com.marcelocantos.pigeon.host"
    compileSdk = 34
    ndkVersion = (
        project.findProperty("pigeon.ndkVersion") as String?
    ) ?: "29.0.14206865"

    defaultConfig {
        applicationId = "com.marcelocantos.pigeon.host"
        minSdk = 33
        targetSdk = 34
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
        testApplicationId = "com.marcelocantos.pigeon.host.test"
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    sourceSets {
        named("androidTest") {
            kotlin.setSrcDirs(listOf("src/androidTest/kotlin"))
        }
    }

    buildTypes {
        debug {
            // No code shrinking on the test host — keep the build fast
            // and the instrumentation target's bytecode untouched.
            isMinifyEnabled = false
        }
    }
}

dependencies {
    implementation(project(":pigeon"))
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")
    androidTestImplementation("androidx.test:runner:1.6.2")
    androidTestImplementation("androidx.test.ext:junit:1.2.1")
    androidTestImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")
}
