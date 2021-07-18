#!/bin/sh

build() {
  now=$(date +'%Y-%m-%d_%T')
  go build -ldflags "-X main.sha1ver=`git rev-parse HEAD` -X main.buildTime=$now -X main.versionTag=0.0.1 -X main.inRelease=1"
  zip -9 binadox-cloud-agent.zip binadox-cloud-agent
}

build
