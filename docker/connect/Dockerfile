FROM debezium/connect:1.7

COPY docker/connect/entrypoint-wrap.sh /entrypoint-wrap.sh
COPY docker/connect/debezium-mysql.json /debezium-mysql.json

CMD ["/entrypoint-wrap.sh"]
