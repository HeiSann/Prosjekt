package comsManager

import (
			"fmt"
			//"time"
		  )




func (fromNet *ComsManager_s)(){
    for{
        msg:=<-fromNet.ExtComs.RecvMsg
    
        switch msg.Msg_type{
		  case "test":
				fmt.Println("tcp msg recieved")
		  case "PING":
		  		fromNet.ExtComs.PingMsg<-msg
		  case "MYCOST":
		  		//go Auction()
		  		//one goroutine for each auction?
		  		//send cost value and ip to the right auction
		  case "NEED COST":
		  		//send cost function and the order it relates to. JSON for order send?? 
		  case "NEWORDER":
		  		//send new order to Order module
		  		//function who unpacks the message and sends it
		  case "ACK":
		  //send to auction so that we are certain that the elevator recieved and saved the order
		  case "SET LIGHT":
		  //Could be in "NEED COST" msg. When other elevator get external orders every elevator needs to set the lights
		  case "DEAD ELEVATOR":
		  //broadcast the ip. Start auctioning the dead elevators external orders
		  case "TCP ERROR"
		  	//go routine
		  	//try to send msg againg
		  	//ig msg fails n times. Send msg to the sender that the msg was lost. Take the order if the msg was an order type
        default:
            fmt.Println("not able to read msg header. Something went terribly wrong, oh god, i have dissapointed the other elevators, they will hate me so much. Pls ctrlC me right now I cant stand this pain any longer")
        }
    }
}
//The auctions will be communicatin to the order module directy nad the select will not be needed? ONLY able to perform one auction at a time?
func (from Order_s)OrderComs(){

for{
        select:
        case msg:=<-ExtComs.ExternalOrder:
			//start goroutine for auction
		case msg:=<-intComs.newOrder:
		case msg:=<-myComst:
		case msg:=<-orderDone:
		case msg:=<-bcastAuction:
		case
}

func (toNet *ComsManager_s)SendMessagesToNet(){

}

func (fromOrder *ComsManager_s)ForwardMessageFromOrder(){
}

func(toOrder *ComsManager_s)DeliverMessageToOrder(){
}
