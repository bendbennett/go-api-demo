FROM confluentinc/cp-schema-registry:7.7.2

COPY docker/schema/entrypoint-wrap.sh /entrypoint-wrap.sh
COPY docker/schema/users-value.json /users-value.json

CMD ["/entrypoint-wrap.sh"]
