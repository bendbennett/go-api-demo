FROM confluentinc/cp-schema-registry

COPY docker/schema/entrypoint-wrap.sh /entrypoint-wrap.sh
COPY docker/schema/users-value.json /users-value.json

CMD ["/entrypoint-wrap.sh"]
