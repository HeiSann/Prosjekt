package comsManager

import (
			"fmt"
			//"time"
		  )




func (fromNet *ComsManager_s)RecieveMessageFromNet(){
    for{
        msg:=<-fromNet.ExtComs.RecvMsg
    
        switch msg.Msg_type{
		  case "test":
				fmt.Println("tcp msg recieved")
		  case "PING":
		  		fromNet.ExtComs.PingMsg<-msg
        default:
            fmt.Println("not able to read msg header")
        }
    }
}

func (toNet *ComsManager_s)SendMessagesToNet(){

}

func (fromOrder *ComsManager_s)ForwardMessageFromOrder(){
}

func(toOrder *ComsManager_s)DeliverMessageToOrder(){
}
