package comsManager

import ("fmt"
		//"time"
		"elevTypes"
		)




func (comsMan *ComsManager_s)RecieveMessageFromNet(){
    for{
        msg:=<-fromNet.ExtComs.RecvMsg
        
    
        switch msg.Msg_type{
		  case "test":
				fmt.Println("tcp msg recieved")
				comsMan.TcpSenderTest(msg.From)
				
		  case "HEARTBEAT":
		  		comsMan.ExtComs.PingMsg<-msg

		  case "COST":
		  		comsMan.intComs.costMsg<-msg

		  case "NEED_COST":
		  		//send cost function and the order it relates to. JSON for order send?? 

		  case "ADD_ORDER":
		  		comsMan.ExtComs.AddOrder<-msg.order

		  case "UPDATE_BACKUP":
				comsMan.ExtComs.UpdateBackup<-msg

		  case "DEAD ELEVATOR":
		  //broadcast the ip. Start auctioning the dead elevators external orders

		  case "TCP ERROR":
		  	//go routine
		  	//try to send msg againg
		  	//ig msg fails n times. Send msg to the sender that the msg was lost. Take the order if the msg was an order type

        default:
            fmt.Println(msg.From)
            fmt.Println("not able to read msg header. Something went terribly wrong, oh god, i have dissapointed the other elevators, they will hate me so much. Pls ctrlC me right now I cant stand this pain any longer")
       }
    }
}

func (toNet *ComsManager_s)TcpSenderTest(to string){
        msg:=elevTypes.Message{}
        msg.From= "129.241.187.156"
        msg.To= to
        msg.Msg_type="test"
        toNet.ExtComs.SendMsg<-msg 
}

/*
func (from Order_s)OrderComs(){

    for{
        select{
        case msg:=<-ExtComs.ExternalOrder:
			//start goroutine for auction
		case msg:=<-intComs.newOrder:
		case msg:=<-myComst:
		case msg:=<-orderDone:
		case msg:=<-bcastAuction:
		}
	}
}
*/
func (toNet *ComsManager_s)SendMessagesToNet(){
}

func (fromOrder *ComsManager_s)ForwardMessageFromOrder(){
}

func(toOrder *ComsManager_s)DeliverMessageToOrder(){
}
