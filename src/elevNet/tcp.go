package elevNet
import(
	"net"
	"fmt"	
	"strings"	
	"message"
)

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
			
		case ip := <-elevNet.ExtComs.ConnectToElev:
			fmt.Println("case connetct to")
			go elevNet.intComs.ConnectElev(ip)
			
		case msg := <-elevNet.ExtComs.SendMsg:
			fmt.Println("case send")
			SendTcpMsg(msg, tcpConnections)
			
			
        case ip := <-elevNet.intComs.dead_elev:
        		fmt.Println("case dead")
            elevNet.intComs.deleteCon(ip, tcpConnections)
			
		}//end select
	}//end for
}

func (toComsMan *ExternalChan_s) listenForTcpMsg (con net.Conn){
	bstream := make([]byte, BUFF_SIZE)
    for {
		_, err := con.Read(bstream[0:])
	    if err!=nil {
			//fmt.Println("error in listen")			
		}else{
			msg:=message.Bytestream2message(bstream)
			toComsMan.RecvMsg<-msg
			
		}
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
   	}
}	

func SendTcpMsg(msg message.Message, tcpConnections map[string]net.Conn){
	ipAddr := msg.To
	bstream:=message.Message2bytestream(msg)
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
	}//end for
}

func (newCon ElevNet_s) registerNewCon(con net.Conn, tcpConnections map[string]net.Conn){ //ta inn conn
	fmt.Println("handle new Con")
	ip:= getConIp(con)

	_, ok := tcpConnections[ip]
	
	if !ok{	
		fmt.Println(ok)
		fmt.Println("connection not in map, adding connection")
		tcpConnections[ip]=con
		go newCon.ExtComs.listenForTcpMsg(con)
		newCon.intComs.newPinger<-ip
	}else{
		fmt.Println("connection already excist")
	}
}

func (toPing *InternalChan_s) deleteCon(ip string, tcpConnections map[string]net.Conn){
    _, ok :=tcpConnections[ip]
    if !ok{
        fmt.Println("connection already lost")
    }else{
        tcpConnections[ip].Close()
        delete(tcpConnections,ip)
        toPing.deadPinger<-ip   
    }
}


func getConIp(con net.Conn)(ip string){
	split:=strings.Split(con.RemoteAddr().String(),":") //splits ip from port
	conIp :=split[0]
	return conIp
	
}
