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


func (elevNet *ElevNet_s)ManageTCPCom(){	
	go elevNet.intComs.listenTcpCon()
	fmt.Println("intcp")

	tcpConnections:= make(map[string]net.Conn)
	
	for {	
		select{
		
		case newTcpCon := <-elevNet.intComs.new_conn:
			fmt.Println("newconn")
			elevNet.registerNewCon(newTcpCon, tcpConnections)
			
		case ip := <-elevNet.intComs.connectToElev:
			fmt.Println("case connetct to")
			go elevNet.intComs.ConnectElev(ip)
			
		case msg := <-elevNet.ExtComs.SendMsg:
			fmt.Println("case send")
			SendTcpMsg(msg, tcpConnections)
			
		case ip := <-elevNet.intComs.deadElev:
        		fmt.Println("case dead")
            deleteCon(ip, tcpConnections)
            
        case msg:=<-elevNet.ExtComs.SendMsgToAll:
        	SendTcpToAll(msg, tcpConnections)
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
			fmt.Println("recieved connection, sending to handle")   			
   		}
	time.Sleep(time.Millisecond*SLEEPTIME)
   	}
}	

func SendTcpMsg(msg elevTypes.Message, tcpConnections map[string]net.Conn){
	ipAddr := msg.To
	bstream, _ := json.Marshal(msg)
	con, ok :=tcpConnections[ipAddr]
	switch ok{
	case true:
		_, err := con.Write(bstream)
		if err!=nil{
			fmt.Println("failed to send msg")
		}else{
			fmt.Println("msg ok")
		}
	case false:
		fmt.Println("error, not a connection")
	}
}	

func SendTcpToAll(msg elevTypes.Message, tcpConnections map[string]net.Conn){//liker ikke helt at elevNet endrer på ipadressen når coms egentlig skal gjørd det?
	for ip, _ := range tcpConnections{
		msg.To=ip
		SendTcpMsg(msg,tcpConnections)
	}
}


func (toManager *InternalChan_s)ConnectElev(ipAdr string){
	atmpts:=0
	for atmpts < CON_ATMPTS{
		serverAddr, err := net.ResolveTCPAddr("tcp",ipAdr+":"+TCP_PORT)
		if err != nil {
			fmt.Println("Error Resolving Address")
			atmpts++
		}else{
			con, err := net.DialTCP("tcp", nil,serverAddr);
			if err != nil {
				fmt.Println("Error DialingTCP")
				atmpts++
			}else{
				fmt.Println("connection ok")
				toManager.new_conn<-con
				fmt.Println("sendt con on chan")
				break
			}
		}//end BIG if/else	
		time.Sleep(time.Millisecond*SLEEPTIME)	
	}//end for
}

func (elevnet ElevNet_s) registerNewCon (con net.Conn, tcpConnections map[string]net.Conn){ //ta inn conn
	fmt.Println("handle new Con")
	ip:= getConIp(con)

	_, ok := tcpConnections[ip]
	
	if !ok{	
		fmt.Println(ok)
		fmt.Println("connection not in map, adding connection")
		tcpConnections[ip]=con
		go elevnet.listenForTcpMsg(con)
		fmt.Println("started to listen")
		//elevnet.intComs.newPinger<-ip
		fmt.Println("send new pinger")
	}else{
		fmt.Println("connection already excist")
	}
}

func deleteCon(ip string, tcpConnections map[string]net.Conn){
    _, ok :=tcpConnections[ip]
    if !ok{
        fmt.Println("connection already lost")
    }else{
        tcpConnections[ip].Close()
        delete(tcpConnections,ip)  
    }
}


func getConIp(con net.Conn)(ip string){
	split:=strings.Split(con.RemoteAddr().String(),":") //splits ip from port
	conIp :=split[0]
	return conIp
	
}
