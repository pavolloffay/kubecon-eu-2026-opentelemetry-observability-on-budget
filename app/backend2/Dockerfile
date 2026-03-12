FROM eclipse-temurin:21-jdk-alpine AS builder

WORKDIR /usr/src/app

COPY gradlew ./
COPY gradle ./gradle
RUN ./gradlew --version --no-daemon

COPY build.gradle settings.gradle ./
RUN ./gradlew dependencies --no-daemon

COPY src ./src
RUN ./gradlew bootJar --no-daemon

FROM eclipse-temurin:21-jre-alpine
WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/build/libs/*.jar app.jar

ENTRYPOINT ["java", "-jar", "./app.jar"]
