#!/bin/bash

bash -c '
hostname=$(hostname -I | awk "{gsub(/ /,\"\"); print $1}")
text_break="\n=============\n"

echo -e "\n${text_break}Waiting for Kafka Connect to start listening on ${hostname}${text_break}"

while [ $(curl -s -o /dev/null -w %{http_code} http://${hostname}:8083/connectors) -ne 200 ] ; do
  echo -e "\n${text_break}$(date) Kafka Connect listener HTTP state:${http_code}" \
          $(curl -s -o /dev/null -w %{http_code} http://${hostname}:8083/connectors)" (waiting for 200)${text_break}"
  sleep 5
done

echo -e $(date) "\n${text_break}Kafka Connect is ready! Listener HTTP state:" \
        $(curl -s -o /dev/null -w %{http_code} http://${hostname}:8083/connectors)${text_break}

connector_setup=$(curl -X POST -H "Content-Type: application/json" -d @/debezium-mysql.json http://${hostname}:8083/connectors)

echo -e "\n${text_break}Connector setup: ${connector_setup}${text_break}"
'&

/docker-entrypoint.sh start
