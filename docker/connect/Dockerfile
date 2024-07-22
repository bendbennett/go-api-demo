FROM quay.io/debezium/connect:3.0

COPY docker/connect/debezium-mysql.json /debezium-mysql.json
COPY docker/connect/entrypoint-wrap.sh /entrypoint-wrap.sh

ENV KAFKA_CONNECT_PLUGINS_DIR=$KAFKA_HOME/connect \
    EXTERNAL_LIBS_DIR=$KAFKA_HOME/external_libs \
    CONNECT_PLUGIN_PATH=$KAFKA_CONNECT_PLUGINS_DIR \
    MAVEN_DEP_DESTINATION=$KAFKA_HOME/libs \
    CONFLUENT_VERSION=7.7.1 \
    AVRO_VERSION=1.12.0 \
    GUAVA_VERSION=33.3.1-jre \
    GUAVA_FAILURE_ACCESS_VERSION=1.0.2

RUN docker-maven-download confluent kafka-connect-avro-converter "$CONFLUENT_VERSION" 262d3d4e880e78fa6bb00469fdf10489 && \
    docker-maven-download confluent kafka-connect-avro-data "$CONFLUENT_VERSION" a0b8b5ac8782fc815b059965ae634706 && \
    docker-maven-download confluent kafka-avro-serializer "$CONFLUENT_VERSION" c335c87e25ae347086631092a07d62bd && \
    docker-maven-download confluent kafka-schema-serializer "$CONFLUENT_VERSION" 56c7911fa0561ed839fa8d9c491cc5f4 && \
    docker-maven-download confluent kafka-schema-registry-client "$CONFLUENT_VERSION" bac9144955be1c0c371e6654f6572a35 && \
    docker-maven-download confluent kafka-schema-converter "$CONFLUENT_VERSION" da55aa8de34c16dff1b8537aff958fe2 && \
    docker-maven-download confluent common-config "$CONFLUENT_VERSION" 03052d98b8ff2c72f5f57b45c00becb4 && \
    docker-maven-download confluent common-utils "$CONFLUENT_VERSION" 2ea73bcfa8c0c719d150bfacdd419ae5 && \
    docker-maven-download central org/apache/avro avro "$AVRO_VERSION" 21fa3115ff1dc131ca6432bc73927fa5 && \
    docker-maven-download central com/google/guava guava "$GUAVA_VERSION" 7b7d80d99af4181db55b00dad50a91bb && \
    docker-maven-download central com/google/guava failureaccess "$GUAVA_FAILURE_ACCESS_VERSION" 3f75955b49b6758fd6d1e1bd9bf777b3

CMD ["/entrypoint-wrap.sh"]
