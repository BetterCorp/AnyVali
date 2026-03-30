plugins {
    kotlin("jvm") version "2.3.20"
    kotlin("plugin.serialization") version "2.3.20"
    id("com.vanniktech.maven.publish") version "0.30.0"
    jacoco
}

group = "com.anyvali"
version = "0.0.1"

repositories {
    mavenCentral()
}

dependencies {
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.10.0")

    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:6.0.3")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

tasks.test {
    useJUnitPlatform()
    finalizedBy(tasks.jacocoTestReport)
}

tasks.jacocoTestReport {
    dependsOn(tasks.test)
    reports {
        xml.required.set(true)
        html.required.set(false)
    }
}

kotlin {
    jvmToolchain(21)
}

mavenPublishing {
    publishToMavenCentral(com.vanniktech.maven.publish.SonatypeHost.CENTRAL_PORTAL)
    signAllPublications()

    coordinates("com.anyvali", "anyvali-kotlin", version.toString())

    pom {
        name.set("AnyVali Kotlin SDK")
        description.set("Portable schema validation for Kotlin")
        url.set("https://anyvali.com")

        licenses {
            license {
                name.set("MIT")
                url.set("https://opensource.org/licenses/MIT")
            }
        }

        developers {
            developer {
                name.set("AnyVali Contributors")
                email.set("hello@anyvali.com")
            }
        }

        scm {
            connection.set("scm:git:https://github.com/BetterCorp/AnyVali.git")
            developerConnection.set("scm:git:ssh://git@github.com/BetterCorp/AnyVali.git")
            url.set("https://github.com/BetterCorp/AnyVali")
        }
    }
}
