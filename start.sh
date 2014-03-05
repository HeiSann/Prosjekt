#!/bin/bash
export GOPATH=$(pwd)

go install elevTypes
go install elevDrivers
go install elevOrders
go install elevFSM
go run src/fsmTest.go
