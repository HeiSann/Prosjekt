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
}
    
type InternalChan_s struct{
	connect_to chan bool
	dead_elev chan string
	new_conn chan net.Conn
	send_msg chan message.Message
}

type ElevNet_s struct{
	ExtComs ExternalChan_s
	intComs InternalChan_s
}
	


func Init()ElevNet_s{

	elevNet:=ElevNet_s{}
	elevNet.ExtComs=externalChannelsInit()
	elevNet.intComs=internalChannelsInit()	
	
	return elevNet
}

func externalChannelsInit()ExternalChan_s{
	extChans:=ExternalChan_s{}
	extChans.RecvMsg = make(chan message.Message,255)
	extChans.SendMsg = make(chan message.Message,255)
	extChans.SendBcast = make(chan message.Message,255)
	extChans.ConnectToElev = make(chan string,255)
	return extChans
}


func internalChannelsInit()InternalChan_s{
	internalChan:=InternalChan_s{}
	internalChan.connect_to = make(chan bool, 255)
	internalChan.dead_elev = make(chan string, 255)
	internalChan.new_conn = make(chan net.Conn, 255)
	internalChan.send_msg = make(chan message.Message)
	return internalChan
}



