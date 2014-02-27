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


func ManageTCPCom(){	
	go listenTcpCon()

	for {	
		select{
		case newTcpCon := <-internalChan.new_conn:
			elevNetMaps.registerNewCon(newTcpCon)
		case ip := <-ExternalChan.ConnectToElev:
			ConnectTcp(ip)
		case msg := <-ExternalChan.SendMsg:
			SendTcpMsg(msg)
        case ip := <-internalChan.dead_elev:
            elevNetMaps.deleteCon(ip)
			
		}//end select
	}//end for
}


func listenMsg(con net.Conn){
	bstream := make([]byte, BUFF_SIZE)
    for {
		_, err := con.Read(bstream[0:])
	    if err!=nil {
			//fmt.Println("error in listen")			
		}else{
			msg:=message.Bytestream2message(bstream)
			ExternalChan.RecvMsg<-msg
			
		}
	}
}

func listenTcpCon(){
	localAddr, err := net.ResolveTCPAddr("tcp",":"+TCP_PORT)
	sock, err := net.ListenTCP("tcp", localAddr)
	if err != nil { return }
 
	for{
		con, err := sock.Accept()
		if err != nil {
			return
		}else{
			internalChan.new_conn<-con
			fmt.Println("recieved connection, sending to handle")   			
   		}
   	}
}	

func SendTcpMsg(msg message.Message){
	ipAddr := msg.To
	bstream:=message.Message2bytestream(msg)
	con, ok :=elevNetMaps.TcpConsMap[ipAddr]
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


func ConnectTcp(ipAdr string){
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
				internalChan.new_conn<-con
				break
			}
		}//end BIG if/else		
	}//end for
}

func (elevNetMaps *elevNetMap)registerNewCon(con net.Conn){ //ta inn conn
	fmt.Println("handle new Con")
	ip:= getConIp(con)

	_, ok := elevNetMaps.TcpConsMap[ip]
	
	if !ok{	
		fmt.Println(ok)
		fmt.Println("connection not in map, adding connection")
		elevNetMaps.TcpConsMap[ip]=con
		go listenMsg(con)
	}else{
		fmt.Println("connection already excist")
	}
}

func (elevNetMaps *elevNetMap) deleteCon(ip string){
    _, ok :=elevNetMaps.TcpConsMap[ip]
    if !ok{
        fmt.Println("connection already lost")
    }else{
        elevNetMaps.TcpConsMap[ip].Close()
        delete(elevNetMaps.TcpConsMap,ip)   
    }
}


func getConIp(con net.Conn)(ip string){
	split:=strings.Split(con.RemoteAddr().String(),":") //splits ip from port
	conIp :=split[0]
	return conIp
	
}
