// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

plugins {
    kotlin("jvm") version "2.1.20" apply false
    id("com.android.library") version "8.7.3" apply false
    kotlin("android") version "2.1.20" apply false
}

// Repositories declared in settings.gradle.kts via
// dependencyResolutionManagement; per-project blocks would conflict
// with PREFER_SETTINGS mode.
