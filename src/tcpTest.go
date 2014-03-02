package main

import "runtime"
import "comsManager"
import "elevNet"

import "fmt"
//import "strings"



const TARGET_IP = "129.241.187.142"



func main() {
	elevNet.TcpChanInit()
	elevNet.ElevNetChanInit()
	comsManager.NetChanInit()
	c := make(chan int)
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("hai")
	
	 
	go elevNet.ExtComs.ManageTCPCom()
	go elevNet.ExtComs.ListenToBroadcast()
	go comsManager.DeliverMsg()
	//go comsManager.SendMsg(msg, elevNet.ElevNetChan)
	go elevNet.ListenToBroadcast(elevNet.ElevNetChan)
	//go coms.SendPckgToAll(coms.ComsChan)
	
	msg:=message.ConstructPckg("129.241.187.152","connectTo", "test")
	//for i:=0;i<1;i++{
	//	network.NetChan.SendUDP<-msg
	//}

	<-c
	
	
}
