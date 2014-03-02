package elevNet
import("net")
import("message")


const TARGET_PORT = "20011"
const LISTEN_PORT = "30011"




type ExternalChan_s struct{
	RecvMsg chan message.Message
	SendMsg chan message.Message  
	SendBcast chan message.Message
	ConnectToElev chan string
	PingMsg chan message.Message
}
    
type InternalChan_s struct{
	connect_to chan bool
	dead_elev chan string
	new_conn chan net.Conn
	send_msg chan message.Message
	timerOut chan bool
	newPinger chan string	
	deadElev chan string
	deadPinger chan string
}

type ElevNet_s struct{
	ip string
	ExtComs ExternalChan_s
	intComs InternalChan_s
	
}
	


func Init()ElevNet_s{
	elevNet:=ElevNet_s{}
	elevNet.ip=GetMyIP()
	elevNet.ExtComs=ExternalChannelsInit()
	elevNet.intComs=InternalChannelsInit()	
	
	return elevNet
}

func ExternalChannelsInit()ExternalChan_s{
	extChans:=ExternalChan_s{}
	extChans.RecvMsg = make(chan message.Message)
	extChans.SendMsg = make(chan message.Message)
	extChans.SendBcast = make(chan message.Message)
	extChans.ConnectToElev = make(chan string)
	extChans.PingMsg = make(chan message.Message)
	return extChans
}


func InternalChannelsInit()InternalChan_s{
	internalChan:=InternalChan_s{}
	internalChan.connect_to = make(chan bool)
	internalChan.dead_elev = make(chan string)
	internalChan.new_conn = make(chan net.Conn)
	internalChan.send_msg = make(chan message.Message)
	internalChan.timerOut = make(chan bool)
	internalChan.newPinger = make(chan string)
	internalChan.deadElev = make(chan string)
	internalChan.deadPinger = make(chan string)
	return internalChan
}



