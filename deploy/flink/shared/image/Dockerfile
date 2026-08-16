# syntax=docker/dockerfile:1.7

FROM flink:2.2.1-java17@sha256:60f0c1ddd87b270cc248f75b5501673de2a731ff321dc8e00329e0d27dc62425

LABEL org.opencontainers.image.title="Apache Flink runtime" \
      org.opencontainers.image.version="2.2.1" \
      org.opencontainers.image.base.name="docker.io/library/flink:2.2.1-java17" \
      org.apache.flink.kafka.connector.version="5.0.0-2.2"

ADD --checksum=sha256:5605c691d11a501382c383fecba37a7a552467da5ab7ba904ef5d6f3d62c5616 \
    --chown=flink:flink \
    https://repo.maven.apache.org/maven2/org/apache/flink/flink-sql-connector-kafka/5.0.0-2.2/flink-sql-connector-kafka-5.0.0-2.2.jar \
    /opt/flink/lib/flink-sql-connector-kafka-5.0.0-2.2.jar

# The official distribution already contains this self-contained filesystem JAR.
# The image entrypoint activates it for every Flink process, including SQL Gateway.
ENV ENABLE_BUILT_IN_PLUGINS=flink-s3-fs-presto-2.2.1.jar

USER root

# Keep only the filesystem implementation used by this runtime. The wildcard is
# intentionally limited to optional filesystem JARs from the base distribution.
RUN find /opt/flink/opt -maxdepth 1 -type f -name '*-fs-had?op-*.jar' -delete

COPY certs/YandexCloudCA.crt certs/YandexInternalRootCA.crt \
    /usr/local/share/ca-certificates/Yandex/

RUN update-ca-certificates && \
    keytool -importcert -noprompt -trustcacerts \
      -alias yandex-cloud-ca \
      -file /usr/local/share/ca-certificates/Yandex/YandexCloudCA.crt \
      -keystore "${JAVA_HOME}/lib/security/cacerts" \
      -storepass changeit && \
    keytool -importcert -noprompt -trustcacerts \
      -alias yandex-internal-root-ca \
      -file /usr/local/share/ca-certificates/Yandex/YandexInternalRootCA.crt \
      -keystore "${JAVA_HOME}/lib/security/cacerts" \
      -storepass changeit

USER flink
