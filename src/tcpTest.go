package main

import ("runtime"
		  "comsManager"
		  "elevNet"	
		  "fmt"
		  	"time"
		  )
		  
		  
		    
		  
		  
		  
func main(){
	c := make(chan int)	
	runtime.GOMAXPROCS(runtime.NumCPU())

	net_s:=elevNet.Init()
	comsMan_s:=comsManager.Init(net_s.Ip,net_s.ExtComs)
	time.Sleep(time.Second)
	
	
	
	fmt.Println("hai")
	
	 
	go net_s.ManageTCPCom()
	go net_s.ListenToBroadcast()
	go comsMan_s.RecieveMessageFromNet()
	go net_s.SendMsgToAll()
	go net_s.RefreshNetwork()
	//go comsManager.SendMsg(msg, elevNet.ElevNetChan)
	//go coms.SendPckgToAll(coms.ComsChan)
	time.Sleep(time.Second)
	comsMan_s.TcpSenderTest("129.241.187.158")
	
	
	<-c
	

}
