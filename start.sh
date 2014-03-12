#!/bin/bash
export GOPATH=$(pwd)

go install elevTypes
go install elevNet
go install comsManager
go install elevDrivers
go install elevOrders
go install elevFSM
go run src/main2.go
