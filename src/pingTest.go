package main

import ("runtime"
		  "comsManager"
		  "elevNet"	
		  "fmt"
		  	"message"
		  	"time"
		  )
		  
		  
		    
		  
		  
		  
func main(){
	c := make(chan int)	
	runtime.GOMAXPROCS(runtime.NumCPU())
	net_s:=elevNet.Init()
	time.Sleep(time.Second)
	
	
	
	fmt.Println("hai")
	
	 
	go net_s.ManageTCPCom()
	go net_s.ExtComs.ListenToBroadcast()
	go comsManager.DeliverMsg(net_s.ExtComs)
	go net_s.ExtComs.SendMsgToAll()
	go net_s.RefreshNetwork()
	//go comsManager.SendMsg(msg, elevNet.ElevNetChan)
	//go coms.SendPckgToAll(coms.ComsChan)
	
	msg:=message.ConstructMessage("129.241.187.255","129.241.187.152","connectTo", "test")
	for i:=0;i<1;i++{
		fmt.Println("yo")
		time.Sleep(time.Second)
		net_s.ExtComs.SendBcast<-msg
		
	}

	<-c
	

}
