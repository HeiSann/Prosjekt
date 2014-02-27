package elevNet
import("net")
import("message")
import("time")

const TARGET_PORT = "20011"
const LISTEN_PORT = "30011"

type ElevNetMap struct{
    TcpConsMap map[string]net.Conn
    PingTimeMap map[string]time.Time
}
var elevNetMaps ElevNetMap

func (elevNetMaps *ElevNetMap) init(){
    elevNetMaps.TcpConsMap = make(map[string]net.Conn)
    elevNetMaps.PingTimeMap =make(map[string]time.Time)
}


type ElevNetChannels struct{
	RecvMsg chan message.Message
	SendMsg chan message.Message  
	SendBcast chan message.Message
	ConnectToElev chan string
}
    
type internalChannels struct{
	connect_to chan bool
	dead_elev chan string
	new_conn chan net.Conn
	send_msg chan message.Message
}
	
var internalChan internalChannels
var ExternalChan ElevNetChannels

func (ExternalChan *ElevNetChannels)Init(){
	ExternalChan.RecvMsg = make(chan message.Message,255)
	ExternalChan.SendMsg = make(chan message.Message,255)
	ExternalChan.SendBcast = make(chan message.Message,255)
	ExternalChan.ConnectToElev = make(chan string,255)

}



func (internalChan *internalChannels) NetChanInit(){
	internalChan.connect_to = make(chan bool, 255)
	internalChan.dead_elev = make(chan string, 255)
	internalChan.new_conn = make(chan net.Conn, 255)
	internalChan.send_msg = make(chan message.Message)
}
