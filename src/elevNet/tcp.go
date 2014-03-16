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
	go elevNet.intComs.listenTcpCon()

	tcpConnections:= make(map[string]net.Conn)
	fmt.Println("ManageTCPCom channe: ",elevNet.ExtComs.SendMsgToAll)
	for {	
		select{		
		
		case newTcpCon := <-elevNet.intComs.new_conn:
			fmt.Println("ManageTCPCom :newconn")
			elevNet.registerNewCon(newTcpCon, tcpConnections)
			
		case ip := <-elevNet.intComs.connectToElev:
			fmt.Println("ManageTCPCom:case connetct to")
			go elevNet.intComs.ConnectElev(ip)
			
		case msg := <-elevNet.ExtComs.SendMsg:
			fmt.Println("ManageTcpCom: case send")
			go elevNet.SendTcpMsg(msg, tcpConnections)
			
		case ip := <-elevNet.intComs.deadElev:
        		fmt.Println("ManageTCPCom:case dead")
            deleteCon(ip, tcpConnections)
            
        case msg:=<-elevNet.ExtComs.SendMsgToAll:
        	fmt.Println("ManageTCPCom:case sendMsgToAll")
        	elevNet.SendTcpToAll(msg, tcpConnections)
		default:
			time.Sleep(time.Millisecond*SLEEPTIME)
			
		}//end select
	}//end for
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

func (toManager *InternalChan_s)listenTcpCon(){
	localAddr, err := net.ResolveTCPAddr("tcp",":"+TCP_PORT)
	sock, err := net.ListenTCP("tcp", localAddr)
	if err != nil { return }
 
	for{
		con, err := sock.Accept()
		if err != nil {
			return
		}else{
			toManager.new_conn<-con
			fmt.Println("listenTcpCon: recieved connection, sending to handle")   			
   		}
	time.Sleep(time.Millisecond*SLEEPTIME)
   	}
}	

func (self *ElevNet_s)SendTcpMsg(msg elevTypes.Message, tcpConnections map[string]net.Conn){
	ipAddr := msg.To
	bstream, _ := json.Marshal(msg)
	con, ok :=tcpConnections[ipAddr]
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
			break
			}
		go self.reConnectAndSend(msg, tcpConnections)
		}
	case false:
		fmt.Println("error, not a connection, trying to connect")
		//toManager.connectToElev<-msg.To
		go self.reConnectAndSend(msg, tcpConnections)		
	}
}	

func (self *ElevNet_s)SendTcpToAll(msg elevTypes.Message, tcpConnections map[string]net.Conn){
	for ip, _ := range tcpConnections{
		msg.To=ip
		fmt.Println("SendTcpToAll:SendTcpToAll:",ip)
		self.SendTcpMsg(msg,tcpConnections)
	}
}


func (toManager *InternalChan_s)ConnectElev(ipAdr string){
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
				toManager.new_conn<-con
				fmt.Println("ConnectElev:sendt con on chan")
				break
			}
		}//end BIG if/else	
		time.Sleep(time.Millisecond*SLEEPTIME)	
	}//end for
}

func (elevnet *ElevNet_s) registerNewCon (con net.Conn, tcpConnections map[string]net.Conn){ //ta inn conn
	fmt.Println("registerNewCon: handle new Con")
	ip:= getConIp(con)

	_, ok := tcpConnections[ip]
	
	if !ok{	
		fmt.Println(ok)
		fmt.Println("registerNewCon:connection not in map, adding connection")
		tcpConnections[ip]=con
		go elevnet.listenForTcpMsg(con)
		fmt.Println("registerNewCon:started to listen")
		//elevnet.intComs.newPinger<-ip
		//fmt.Println("registerNewCon:send new pinger")
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
	split:=strings.Split(con.RemoteAddr().String(),":") //splits ip-part from port
	conIp :=split[0]
	return conIp
	
}

func (elevNet *ElevNet_s)reConnectAndSend(msg elevTypes.Message, tcpMap map[string]net.Conn){
	elevNet.intComs.ConnectElev(msg.To)
	ipAddr := msg.To
	bstream, _ := json.Marshal(msg)
	con, ok :=tcpMap[ipAddr]
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
			break
			}
		}
		if msg.Type=="ADD_ORDER"{
			elevNet.ExtComs.RecvMsg<-msg
			fmt.Println("reConnectAndSend:error int send msg. Taking order self")
		}
		
	case false:
		if msg.Type=="ADD_ORDER"{
			elevNet.ExtComs.RecvMsg<-msg
			fmt.Println("reConnectAndSend:error in connection, reConnectFailed, taking order self")
			//send addOrder msg back to elevator !!!
		}
	}
}
