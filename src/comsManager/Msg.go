package comsManager

import "elevNet"
import "fmt"
import "message"




func DeliverMsg(fromNet elevNet.ExternalChan_s){
    for{
        msg:=<-fromNet.RecvMsg
    
        switch msg.Msg_type{
        case "connectTo":
            fmt.Println("The msg is of type udp")
				fromNet.ConnectToElev<-msg.From
		  case "test":
				fmt.Println("tcp msg recieved")
		  case "PING":
		  		fromNet.PingMsg<-msg
        default:
            fmt.Println("not able to read msg header")
        }
    }
}

func MsgSend(msg message.Message, toNet elevNet.ExternalChan_s){ //TTEST
	for{
		select{
		case <-NetChan.SendUDP:
			toNet.SendBcast<-msg
		}
	}
}


