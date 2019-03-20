#!/bin/bash

curl -X DELETE http://localhost:18800/tamper -H 'Api-Key: blah' -d \
'{"digest":"'$1'" , "value": "'$2'"}'

