#!/bin/bash

export GOPATH=$PWD/gopath
export PATH=$GOPATH/bin:$PATH

cd gopath/src/github.com/cloudfoundry-incubator/spiff
export GOPATH=${PWD}/Godeps/_workspace:$GOPATH
export PATH=${PWD}/Godeps/_workspace/bin:$PATH

go build -o ci/spiff .
