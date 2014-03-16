package comsManager

import ("fmt"
		"time"
		"elevTypes"
		"strconv"
		)




func (comsMan *ComsManager_s)RecieveMessageFromNet(){
    for{
        msg:=<-comsMan.ExtComs.RecvMsg
        
    
        switch msg.Type{
				
		  case "PING":
		  		comsMan.ExtComs.PingMsg<-msg

		  case "COST":
		  		comsMan.intComs.costMsg<-msg
				fmt.Println("\t RecieveMessegeFromNet: COST;", msg.Payload)
				
		  case "NEED_COST":
		  		cost :=comsMan.getMyCost(msg.Order) //remember if only cost<cost
		  		costMsg:=constructCostMsg(comsMan.Ip, msg.From, msg.Order, cost)
		  		comsMan.ExtComs.SendMsg<-costMsg
		  		fmt.Println("\t sendt my cost to the elevator requiring it, COST=", costMsg.Payload)		  		

		  case "ADD_ORDER":
		  		comsMan.ExtComs.AddOrder<-msg.Order
		  		fmt.Println("\t RecieveMessegeFromNet: ADD_ORDER:", msg.Order)

		  case "UPDATE_BACKUP":
				comsMan.ExtComs.RecvOrderUpdate<-msg
				fmt.Println("\t RecieveMessegeFromNet: UPDATE_BACKUP with order;", msg.Order)

		  case "DEAD ELEVATOR":
		  //broadcast the ip. Start auctioning the dead elevators external orders

		  case "TCP ERROR":
		  	//go routine
		  	//try to send msg againg
		  	//ig msg fails n times. Send msg to the sender that the msg was lost. Take the order if the msg was an order type

        default:
            fmt.Println("\t", msg.From)
            fmt.Println("\tnot able to read msg header. Something went terribly wrong, oh god, i have dissapointed the other elevators, they will hate me so much. Pls ctrlC me right now I cant stand this pain any longer")
       }
    }
}

func (self *ComsManager_s)TcpSenderTest(to string){
        msg:=elevTypes.Message{}
        msg.From= "129.241.187.156"
        msg.To= to
        msg.Type="test"
        self.ExtComs.SendMsg<-msg 
}


func (self *ComsManager_s)InternalCommunication(){ //change name
	fmt.Println("\tCOMSMANAGER: DEAD ELEV CHANNEL", self.ExtComs.DeadElev)
	for{
		select{
		case order:=<- self.ExtComs.SendOrderUpdate:
			fmt.Println("\t comsManager.ForwardMessageFromOrder: got order: ", order,"trying to send")
			msg:=constructUpdateMsg(self.Ip, order, self.Ip)
			self.ExtComs.SendMsgToAll<-msg
			fmt.Println("\t comsManager.ForwardMessageFromOrder: sendt msg self.ExtComs.SendMsgToAll<-msg, msg=", msg)
			
		case order:=<-self.intComs.needCost:
			needCostMsg:=constructNeedCostMsg(self.Ip, order)
    		fmt.Println("\t comsManager: needcostMsg created. Trying to send")
    		self.ExtComs.SendMsgToAll<-needCostMsg
    		fmt.Println("\t comsManager: send need cost Msg to all tcp elevators")		
    	
    	case deadIp:=<-self.ExtComs.DeadElev:
    		fmt.Println("\t ForwardMsg: dead ip:", deadIp)
			self.ExtComs.AuctionDeadElev<-deadIp		
		default:
			time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}
	}	
}


func constructCostMsg(myIp string, toIp string, order elevTypes.Order_t, cost int)elevTypes.Message{
	msg:=elevTypes.Message{}
	msg.From = myIp
	msg.To = toIp
	msg.Type="COST"
	msg.Payload = strconv.Itoa(cost)
	msg.Order = order	 
	return msg
}

func constructNeedCostMsg(myIP string, order elevTypes.Order_t)elevTypes.Message{
	msg:=elevTypes.Message{}
	msg.From = myIP
	msg.Type = "NEED_COST"
	msg.Order = order
	return msg
}


