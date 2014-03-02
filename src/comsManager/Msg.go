package comsManager

import ("elevNet"
			"fmt"
			"elevTypes"
		  )




func DeliverMsg(fromNet elevNet.ExternalChan_s){
    for{
        msg:=<-fromNet.RecvMsg
    
        switch msg.Msg_type{
		  case "test":
				fmt.Println("tcp msg recieved")
		  case "PING":
		  		fromNet.PingMsg<-msg
        default:
            fmt.Println("not able to read msg header")
        }
    }
}

func MsgSend(msg elevTypes.Message, toNet elevNet.ExternalChan_s){ //TTEST
	for{
		select{
		}
	}
}


