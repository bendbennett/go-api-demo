#!/bin/bash

bash -c '
uri=http://go-api-demo-connect:8083/connectors
text_break="\n=============\n"

echo -e "\n${text_break}Waiting for Kafka Connect to start listening on ${uri}${text_break}"

while [ $(curl -s -o /dev/null -w %{http_code} ${uri}) -ne 200 ] ; do
  echo -e "\n${text_break}$(date) Kafka Connect listener HTTP state: ${http_code}" \
          $(curl -s -o /dev/null -w %{http_code} ${uri})" (waiting for 200)${text_break}"
  sleep 5
done

echo -e $(date) "\n${text_break}Kafka Connect is ready! Listener HTTP state:" \
        $(curl -s -o /dev/null -w %{http_code} ${uri})${text_break}

connector_setup=$(curl -X POST -H "Content-Type: application/json" -d @/debezium-mysql.json ${uri})

echo -e "\n${text_break}Connector setup: ${connector_setup}${text_break}"
'&

/docker-entrypoint.sh start
