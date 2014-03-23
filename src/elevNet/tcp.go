package elevNet
import(
	"net"
	"strings"	
	"elevTypes"
	"time"
	"encoding/json"
)
const SLEEPTIME = 5
const CON_ATMPTS = 10
const TCP_PORT = "30000" //All elevators will listen to this port for TCP connections
const BUFF_SIZE = 1024
const SEND_ATMPTS = 5


func (elevNet *ElevNet_s)ManageTCPCom(){	
	go elevNet.intComs.listenForTcpCon()

	tcpConnections:= make(map[string]net.Conn)
	for {	
		select{		
		
		case newTcpCon := <-elevNet.intComs.newConn:
			elevNet.registerNewCon(newTcpCon, tcpConnections)
			
		case ip := <-elevNet.intComs.connectToElev:
			go elevNet.intComs.connectElev(ip)
			
		case msg := <-elevNet.ExtComs.SendMsg:
			go elevNet.sendTcpMsg(msg, tcpConnections)
			
		case ip := <-elevNet.intComs.deadElev:
			deleteCon(ip, tcpConnections)
			
		case msg:=<-elevNet.ExtComs.SendMsgToAll:
			elevNet.sendTcpToAll(msg, tcpConnections)
		
		case msg:= <-elevNet.intComs.tcpFail:
			elevNet.ExtComs.FailedTcpMsg<-msg
		default:
			time.Sleep(time.Millisecond*SLEEPTIME)
			
		}
	}
}


func (toComsMan *ElevNet_s) listenForTcpMsg (con net.Conn){
	bstream := make([]byte, BUFF_SIZE)
	for {
		n, err := con.Read(bstream[0:])
		if err!=nil {
			//fmt.Println("error in listen")			
		}else{
			var msg elevTypes.Message
			err := json.Unmarshal(bstream[0:n], &msg)
			if err == nil{
			toComsMan.ExtComs.RecvMsg<-msg
			}
	}	
   time.Sleep(time.Millisecond*SLEEPTIME)
	}
}


func (toManager *InternalChan_s) listenForTcpCon(){
	localAddr, err := net.ResolveTCPAddr("tcp",":"+TCP_PORT)
	sock, err := net.ListenTCP("tcp", localAddr)
	if err != nil { return }
 
	for{
		con, err := sock.Accept()
		if err != nil {
			return
		}else{
			toManager.newConn<-con 			
   		}
	time.Sleep(time.Millisecond*SLEEPTIME)
   	}
}	


func (self *ElevNet_s)sendTcpMsg(msg elevTypes.Message, tcpConnections map[string]net.Conn){
	ipAddr := msg.To
	bstream, _ := json.Marshal(msg)
	con, ok :=tcpConnections[ipAddr]
trySend:
	switch ok{
	case true:
		try:=0
		for try < SEND_ATMPTS{
			_, err := con.Write(bstream)
			if err!=nil{
				try=try+1		
			}else{
				break trySend
			}
		}
		go self.reConnectAndSend(msg, tcpConnections)
	case false:
		go self.reConnectAndSend(msg, tcpConnections)		
	}
}	


func (self *ElevNet_s)sendTcpToAll(msg elevTypes.Message, tcpConnections map[string]net.Conn){
	for ip, _ := range tcpConnections{
		msg.To=ip
		self.sendTcpMsg(msg,tcpConnections)
	}
}


func (toManager *InternalChan_s)connectElev(ipAdr string){
	atmpts:=0
	for atmpts < CON_ATMPTS{
		serverAddr, err := net.ResolveTCPAddr("tcp",ipAdr+":"+TCP_PORT)
		if err != nil {
			atmpts++
		}else{
			con, err := net.DialTCP("tcp", nil,serverAddr);
			if err != nil {
				atmpts++
			}else{
				toManager.newConn<-con
				break
			}
		}
		time.Sleep(time.Millisecond*SLEEPTIME)	
	}
}


func (elevnet *ElevNet_s) registerNewCon (con net.Conn, tcpConnections map[string]net.Conn){
	ip:= getConIp(con)

	_, ok := tcpConnections[ip]
	
	if !ok{	
		tcpConnections[ip]=con
		go elevnet.listenForTcpMsg(con)
	}
}


func deleteCon(ip string, tcpConnections map[string]net.Conn){
	_, ok :=tcpConnections[ip]
	if ok{
		tcpConnections[ip].Close()
		delete(tcpConnections,ip)
	}
}


func getConIp(con net.Conn)(ip string){
	//splits ip-part from port
	split:=strings.Split(con.RemoteAddr().String(),":") 
	conIp :=split[0]
	return conIp
	
}


func (elevNet *ElevNet_s)reConnectAndSend(msg elevTypes.Message, tcpMap map[string]net.Conn){
	elevNet.intComs.connectElev(msg.To)
	ipAddr := msg.To
	bstream, _ := json.Marshal(msg)
			
	con, ok :=tcpMap[ipAddr]
trySend:
	switch ok{
	case true:
		try:=0
		for try < SEND_ATMPTS{
			_, err := con.Write(bstream)
			if err!=nil{
				try=try+1		
			}else{
				break trySend
			}
		}
		elevNet.intComs.tcpFail<-msg				
	case false:
			elevNet.intComs.tcpFail<-msg
	}
}
