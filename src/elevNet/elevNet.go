package elevNet
import("net"
		"elevTypes"
		//"strings"
		)


const TARGET_PORT = "20011"
const LISTEN_PORT = "30011"


type ElevNet_s struct{
	Ip string
	ExtComs elevTypes.Net_ExtComs_s
	intComs InternalChan_s
}

type InternalChan_s struct{
	connect_to chan bool
	new_conn chan net.Conn
	send_msg chan elevTypes.Message
	timerOut chan bool
	newPinger chan string	
	deadElev chan string
	deadPinger chan string
	connectToElev chan string
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
	extChans.PingMsg = make(chan elevTypes.Message)
	extChans.SendMsgToAll =make(chan elevTypes.Message)
	return extChans
}


func InternalChannelsInit()InternalChan_s{
	internalChan:=InternalChan_s{}
	internalChan.connect_to = make(chan bool)
	internalChan.new_conn = make(chan net.Conn)
	internalChan.send_msg = make(chan elevTypes.Message)
	internalChan.timerOut = make(chan bool)
	internalChan.newPinger = make(chan string)
	internalChan.deadElev = make(chan string)
	internalChan.deadPinger = make(chan string)
	internalChan.connectToElev =make(chan string)
	return internalChan
}

/*
//not generic, could use reflect...
func message2bytestream (m elevTypes.Message) []byte{
	msg := m.To +"~"+m.From +"~"+ m.Msg_type +"~"+ m.Payload
	return []byte(msg+"\x00")
}

//not generic, could use reflect..
func bytestream2message(m []byte) elevTypes.Message{
	msg_string := string(m[:])
	msg_array := strings.Split(msg_string, "~")
	msg := elevTypes.Message{msg_array[0], msg_array[1], msg_array[2], msg_array[3]}
	return msg
}*/

