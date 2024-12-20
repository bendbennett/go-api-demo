# Debezium connect dockerfile uses .jar(s) that are 
# specific to schema registry version.
# Updates to this image should be undertaken at the 
# same time as .jar file updates.
FROM confluentinc/cp-schema-registry:7.8.0

COPY docker/schema/entrypoint-wrap.sh /entrypoint-wrap.sh
COPY docker/schema/users-value.json /users-value.json

CMD ["/entrypoint-wrap.sh"]
