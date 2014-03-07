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
		  		//goroutine need to end when a winner is selected and the auctioneer has recieved an acc
		  case "NEED COST"
		  		//send cost function and the order it relates to. JSON for order send?? 
		  case "NEWORDER"
		  		//send new order to Order module
		  		//function who unpacks the message and sends it
        default:
            fmt.Println("not able to read msg header. Something went terribly wrong, oh god, i have dissapointed the other elevators, they will hate me so much")
        }
    }
}

func (from Order_s)OrderComs

func (toNet *ComsManager_s)SendMessagesToNet(){

}

func (fromOrder *ComsManager_s)ForwardMessageFromOrder(){
}

func(toOrder *ComsManager_s)DeliverMessageToOrder(){
}
