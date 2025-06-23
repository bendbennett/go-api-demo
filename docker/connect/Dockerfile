FROM quay.io/debezium/connect:3.1.2.Final

COPY docker/connect/debezium-mysql.json /debezium-mysql.json
COPY docker/connect/entrypoint-wrap.sh /entrypoint-wrap.sh

ENV KAFKA_CONNECT_PLUGINS_DIR=$KAFKA_HOME/connect \
    EXTERNAL_LIBS_DIR=$KAFKA_HOME/external_libs \
    CONNECT_PLUGIN_PATH=$KAFKA_CONNECT_PLUGINS_DIR \
    MAVEN_DEP_DESTINATION=$KAFKA_HOME/libs \
    CONFLUENT_VERSION=7.9.0 \
    AVRO_VERSION=1.12.0 \
    GUAVA_VERSION=33.4.8-jre \
    GUAVA_FAILURE_ACCESS_VERSION=1.0.3

RUN docker-maven-download confluent kafka-connect-avro-converter "$CONFLUENT_VERSION" 183e0199fb85d24c0bcca22c1ede511e && \
    docker-maven-download confluent kafka-connect-avro-data "$CONFLUENT_VERSION" 841e68574e9eb552a4f2752975267a22 && \
    docker-maven-download confluent kafka-avro-serializer "$CONFLUENT_VERSION" d77ccb260ac3a2b8bc322bd2d2ea70c2 && \
    docker-maven-download confluent kafka-schema-serializer "$CONFLUENT_VERSION" 2d9890f32811c6adc47c48cf91300384 && \
    docker-maven-download confluent kafka-schema-registry-client "$CONFLUENT_VERSION" cd0bdd273f4fbca342be26361cf8934e && \
    docker-maven-download confluent kafka-schema-converter "$CONFLUENT_VERSION" e97a84709b2518ab97f83596e48c2c37 && \
    docker-maven-download confluent common-config "$CONFLUENT_VERSION" 9bcdf1e53f73f690165c1800a81c05b6 && \
    docker-maven-download confluent common-utils "$CONFLUENT_VERSION" 8cf16e6afc4b95ab070128112932bcfb && \
    docker-maven-download central org/apache/avro avro "$AVRO_VERSION" 21fa3115ff1dc131ca6432bc73927fa5 && \
    docker-maven-download central com/google/guava guava "$GUAVA_VERSION" 72920caab34426c5815e3b00f80e3b01 && \
    docker-maven-download central com/google/guava failureaccess "$GUAVA_FAILURE_ACCESS_VERSION" 29a782e90f6b37218b18bb880d2a8f4a

CMD ["/entrypoint-wrap.sh"]
