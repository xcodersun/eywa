Roadmap:
Support tagging connections
Support pub/sub
Configurate message schema, current uuid v4 - url base64 encoding?
Middlewares that detects the location by IP address(enterprise version)



API's:

curl -vvv  "localhost:8080/channels"

curl -vvv  -XDELETE "localhost:8080/channels/test"

curl -XPUT -vvv  "localhost:8080/channels/test" -d '{
  "name":"test",
  "description": "test desc  IOO",
  "tags": ["tag1", "tag2"],
  "fields": {
    "temp": "float"
  }
}'
