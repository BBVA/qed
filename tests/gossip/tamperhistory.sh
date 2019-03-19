#!/bin/bash

curl -X PATCH http://localhost:18800/tamper -H 'Api-Key: blah' -d \
'{"digest":"$1", "value":"$2"}'

