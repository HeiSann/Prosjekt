package elevNet

import (
	"net"	
	"encoding/json"
	"elevTypes"
)

const UDP_PORT ="20000" //All Elevs send and listen to this Broadcast Port

func (fromComs *ElevNet_s) SendMsgToAll(){
	bcastIP:=GetBroadcastIP(GetMyIP())
	
	serverAddr, err := net.ResolveUDPAddr("udp",bcastIP+":"+UDP_PORT)
	if err != nil {return}

	con, err := net.DialUDP("udp", nil, serverAddr)	
	if err != nil {return}
	
	for {
		msg:=<-fromComs.ExtComs.SendBcast
		b, _ := json.Marshal(msg)
		con.Write(b)
	}		
}

func (toComs *ElevNet_s)ListenToBroadcast() {
	myIp :=GetMyIP()
	bcastIP:=GetBroadcastIP(myIp)
	

	serverAddr, err := net.ResolveUDPAddr("udp",bcastIP+":"+UDP_PORT)
	if err != nil { return }
	
	psock, err := net.ListenUDP("udp4", serverAddr)	
	if err != nil { return }
	
	buf := make([]byte,255)
 
  	for {
  		n, remoteAddr, err := psock.ReadFromUDP(buf[0:])
			if err != nil { return }
		
			if remoteAddr.IP.String() != myIp {
					var msg elevTypes.Message
					err := json.Unmarshal(buf[0:n], &msg)
					if err ==nil{
						toComs.ExtComs.RecvMsg<-msg
					}
			}	   
	  }	
}

