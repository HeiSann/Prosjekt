package elevNet
import("net"
		"elevTypes"
		)


const TARGET_PORT = "20011"
const LISTEN_PORT = "30011"


type ElevNet_s struct{
	Ip string
	ExtComs elevTypes.Net_ExtComs_s
	intComs InternalChan_s
}

type InternalChan_s struct{
	newConn chan net.Conn
	timerOut chan bool
	newHeartbeater chan string	
	deadElev chan string
	deadHeartbeater chan string
	connectToElev chan string
	tcpFail chan elevTypes.Message
	
}


func Init()ElevNet_s{
	elevNet:=ElevNet_s{}
	elevNet.Ip=GetMyIP()
	elevNet.ExtComs=ExternalChannelsInit()
	elevNet.intComs=InternalChannelsInit()	
	
	go elevNet.ListenToBroadcast()
	go elevNet.ManageTCPCom()
	go elevNet.RefreshNetwork()
	go elevNet.SendMsgToAll()
	return elevNet
}

func ExternalChannelsInit() elevTypes.Net_ExtComs_s{
	extChans:=elevTypes.Net_ExtComs_s{}
	extChans.RecvMsg = make(chan elevTypes.Message)
	extChans.SendMsg = make(chan elevTypes.Message)	
	extChans.SendBcast = make(chan elevTypes.Message)
	extChans.HeartbeatMsg = make(chan elevTypes.Message)
	extChans.SendMsgToAll =make(chan elevTypes.Message)
	extChans.DeadElev =make(chan string)
	extChans.NewElev = make(chan string)
	extChans.FailedTcpMsg = make(chan elevTypes.Message)
	return extChans
}


func InternalChannelsInit()InternalChan_s{
	internalChan:=InternalChan_s{}
	internalChan.newConn = make(chan net.Conn)
	internalChan.timerOut = make(chan bool)
	internalChan.newHeartbeater = make(chan string)
	internalChan.deadElev = make(chan string)
	internalChan.deadHeartbeater = make(chan string)
	internalChan.connectToElev =make(chan string)
	internalChan.tcpFail = make(chan elevTypes.Message)
	return internalChan
}


