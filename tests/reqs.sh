#!/bin/bash

curl -X POST http://localhost:8080/addstreamurl \
     -H "Content-Type: application/json" \
     -d '{"url": "http://example.com/"}'

curl -X POST http://localhost:8080/addstreamurl \
     -H "Content-Type: application/json" \
     -d '{"url": "https://google.com/"}'

curl -X POST http://localhost:8080/addstreamurl \
     -H "Content-Type: application/json" \
     -d '{"url": "https://panderu.org/"}'

curl http://localhost:8080/print


