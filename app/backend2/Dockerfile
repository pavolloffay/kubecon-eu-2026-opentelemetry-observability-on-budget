FROM gradle:8.4-jdk21 AS builder

WORKDIR /usr/src/app

COPY build.gradle settings.gradle ./
COPY gradle ./gradle
RUN gradle dependencies --no-daemon

COPY src ./src
RUN gradle bootJar --no-daemon

FROM eclipse-temurin:21-jre-alpine
WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/build/libs/*.jar app.jar

ADD https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar javaagent.jar

COPY run.sh .

ENTRYPOINT ["./run.sh"]
