#!/bin/sh

build() {
  now=$(date +'%Y-%m-%d_%T')
  go build -ldflags "-X main.sha1ver=`git rev-parse HEAD` -X main.buildTime=$now"
  zip -9 binadox-cloud-agent.zip binadox-cloud-agent
}

build
