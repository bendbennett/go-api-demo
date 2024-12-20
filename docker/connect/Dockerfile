FROM quay.io/debezium/connect:3.0.6.Final

COPY docker/connect/debezium-mysql.json /debezium-mysql.json
COPY docker/connect/entrypoint-wrap.sh /entrypoint-wrap.sh

ENV KAFKA_CONNECT_PLUGINS_DIR=$KAFKA_HOME/connect \
    EXTERNAL_LIBS_DIR=$KAFKA_HOME/external_libs \
    CONNECT_PLUGIN_PATH=$KAFKA_CONNECT_PLUGINS_DIR \
    MAVEN_DEP_DESTINATION=$KAFKA_HOME/libs \
    CONFLUENT_VERSION=7.8.0 \
    AVRO_VERSION=1.12.0 \
    GUAVA_VERSION=33.4.0-jre \
    GUAVA_FAILURE_ACCESS_VERSION=1.0.2

RUN docker-maven-download confluent kafka-connect-avro-converter "$CONFLUENT_VERSION" 71a28b25ec993b47c07a0579b9a74ec0 && \
    docker-maven-download confluent kafka-connect-avro-data "$CONFLUENT_VERSION" 10fdfdd7e9370df5d9ff63411dab0a4e && \
    docker-maven-download confluent kafka-avro-serializer "$CONFLUENT_VERSION" 54d557c0df44a068d6c80abf81d3f36d && \
    docker-maven-download confluent kafka-schema-serializer "$CONFLUENT_VERSION" 88a1ff2191befa4ad0a7cfbdce2607fa && \
    docker-maven-download confluent kafka-schema-registry-client "$CONFLUENT_VERSION" 719683762f460f186b8426780b7594cb && \
    docker-maven-download confluent kafka-schema-converter "$CONFLUENT_VERSION" 0fe708e96598cf6317b8992a65087069 && \
    docker-maven-download confluent common-config "$CONFLUENT_VERSION" 182c208c59631808eba3f3426552e547 && \
    docker-maven-download confluent common-utils "$CONFLUENT_VERSION" 684c41c3f9d78fbccb3d372c8504cfc0 && \
    docker-maven-download central org/apache/avro avro "$AVRO_VERSION" 21fa3115ff1dc131ca6432bc73927fa5 && \
    docker-maven-download central com/google/guava guava "$GUAVA_VERSION" 5732af16367192820c7bf177e9b29512 && \
    docker-maven-download central com/google/guava failureaccess "$GUAVA_FAILURE_ACCESS_VERSION" 3f75955b49b6758fd6d1e1bd9bf777b3

CMD ["/entrypoint-wrap.sh"]
