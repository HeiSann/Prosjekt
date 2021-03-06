package elevNet
import(
	"net"
	"fmt"	
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
	fmt.Println("go tcp manager")
	go elevNet.intComs.listenForTcpCon()

	tcpConnections:= make(map[string]net.Conn)
	fmt.Println("ManageTCPCom channe: ",elevNet.ExtComs.SendMsgToAll)
	for {	
		select{		
		
		case newTcpCon := <-elevNet.intComs.newConn:
			fmt.Println("ManageTCPCom :newconn")
			elevNet.registerNewCon(newTcpCon, tcpConnections)
			
		case ip := <-elevNet.intComs.connectToElev:
			fmt.Println("ManageTCPCom:case connetct to")
			go elevNet.intComs.connectElev(ip)
			
		case msg := <-elevNet.ExtComs.SendMsg:
			fmt.Println("ManageTcpCom: case send")
			go elevNet.sendTcpMsg(msg, tcpConnections)
			
		case ip := <-elevNet.intComs.deadElev:
				fmt.Println("ManageTCPCom:case dead")
			deleteCon(ip, tcpConnections)
			
		case msg:=<-elevNet.ExtComs.SendMsgToAll:
			fmt.Println("ManageTCPCom:case sendMsgToAll")
			elevNet.sendTcpToAll(msg, tcpConnections)
		
		case msg:= <-elevNet.intComs.tcpFail:
			elevNet.ExtComs.FailedTcpMsg<-msg
			fmt.Println("ManageTCPCom:case tcpFail")
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
			fmt.Println("listenTcpCon: recieved connection, sending to handle")   			
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
				fmt.Println("SendTcpMsg:failed to send msg")
				try=try+1		
			}else{
				fmt.Println("SendTcpMsg: msg ok")
				break trySend
			}
		}
		go self.reConnectAndSend(msg, tcpConnections)
	case false:
		fmt.Println("error, not a connection, trying to connect")
		go self.reConnectAndSend(msg, tcpConnections)		
	}
}	


func (self *ElevNet_s)sendTcpToAll(msg elevTypes.Message, tcpConnections map[string]net.Conn){
	for ip, _ := range tcpConnections{
		msg.To=ip
		fmt.Println("SendTcpToAll:SendTcpToAll:",ip)
		self.sendTcpMsg(msg,tcpConnections)
	}
}


func (toManager *InternalChan_s)connectElev(ipAdr string){
	atmpts:=0
	for atmpts < CON_ATMPTS{
		serverAddr, err := net.ResolveTCPAddr("tcp",ipAdr+":"+TCP_PORT)
		if err != nil {
			fmt.Println("ConnectElev:Error Resolving Address")
			atmpts++
		}else{
			con, err := net.DialTCP("tcp", nil,serverAddr);
			if err != nil {
				fmt.Println("ConnectElev:Error DialingTCP")
				atmpts++
			}else{
				fmt.Println("ConnectElev:connection ok")
				toManager.newConn<-con
				fmt.Println("ConnectElev:sendt con on chan")
				break
			}
		}
		time.Sleep(time.Millisecond*SLEEPTIME)	
	}
}


func (elevnet *ElevNet_s) registerNewCon (con net.Conn, tcpConnections map[string]net.Conn){
	fmt.Println("registerNewCon: handle new Con")
	ip:= getConIp(con)

	_, ok := tcpConnections[ip]
	
	if !ok{	
		fmt.Println(ok)
		fmt.Println("registerNewCon:connection not in map, adding connection")
		tcpConnections[ip]=con
		go elevnet.listenForTcpMsg(con)
		fmt.Println("registerNewCon:started to listen")
	}else{
		fmt.Println("registerNewCon:connection already excist")
	}
}


func deleteCon(ip string, tcpConnections map[string]net.Conn){
	_, ok :=tcpConnections[ip]
	if !ok{
		fmt.Println("deleteCon:connection already lost")
	}else{
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
	bstream, err0 := json.Marshal(msg)
	fmt.Println("reConectAndSend Json:", err0)
	//json fail?
			
	con, ok :=tcpMap[ipAddr]
trySend:
	switch ok{
	case true:
		try:=0
		for try < SEND_ATMPTS{
			_, err := con.Write(bstream)
			if err!=nil{
				fmt.Println("reConnectAndSend:failed to send msg")
				try=try+1		
			}else{
				fmt.Println("reConnectAndSend: msg ok")
				break trySend
			}
		}
		elevNet.intComs.tcpFail<-msg				
	case false:
			elevNet.intComs.tcpFail<-msg
			fmt.Println("reConnectAndSend:error in connection, reConnectFailed, taking order self")
	}
}
