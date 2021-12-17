#!/bin/bash

bash -c '
domain=http://go-api-demo-schema-registry:8081
endpoint=/subjects/go_api_demo_db.go-api-demo.users-value/versions

text_break="\n=============\n"

echo -e "\n${text_break}Waiting for Schema Registry to start listening on ${domain}${text_break}"

while [ $(curl -s -o /dev/null -w %{http_code} ${domain}) -ne 200 ] ; do
  echo -e "\n${text_break}$(date) Schema Registry listener HTTP state: ${http_code}" \
          $(curl -s -o /dev/null -w %{http_code} ${domain})" (waiting for 200)${text_break}"
  sleep 5
done

echo -e $(date) "\n${text_break}Schema Registry is ready! Listener HTTP state:" \
        $(curl -s -o /dev/null -w %{http_code} ${domain})${text_break}

registry_setup=$(curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" -d @/users-value.json ${domain}${endpoint})

echo -e "\n${text_break}Schema Registry setup: ${registry_setup}${text_break}"
'&

/etc/confluent/docker/run
