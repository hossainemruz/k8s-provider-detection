#!/bin/bash
set -x

go build -o provider-detector main.go

chmod +x ./provider-detector

cat >Dockerfile <<EOL
FROM busybox:glibc

COPY ./provider-detector /usr/bin/provider-detector

ENTRYPOINT ["/usr/bin/provider-detector"]
EOL

docker build -t emruzhossain/k8s-provider-detector .
docker push emruzhossain/k8s-provider-detector
rm provider-detector Dockerfile
