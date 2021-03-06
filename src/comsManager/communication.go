package comsManager

import ("fmt"
		"time"
		"elevTypes"
		)


func (comsMan *ComsManager_s)RecieveMessageFromNet(){
	for{
		msg:=<-comsMan.ExtComs.RecvMsg
		
		switch msg.Type{
				
		  case "HEARTBEAT":
		  		comsMan.ExtComs.HeartbeatMsg<-msg

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

		default:
			fmt.Println("\t", msg.From)
			fmt.Println("\tnot able to read msg header. Something went terribly wrong, oh god, i have dissapointed the other elevators, they will hate me so much. Pls ctrlC me right now I cant stand this pain any longer")
	   }
	}
}


func (self *ComsManager_s)ManageCommunicationFromNetAndOrder(){ 
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

		case newIp :=<-self.ExtComs.NewElev:
			newElevUpdate:=constructNewOrderMsg(newIp, self.Ip, elevTypes.Order_t{})
			self.ExtComs.CheckNewElev<-newElevUpdate
			fmt.Println("\t InternalCommunication: commanded order to check if new elevator has any inside orders:", newIp)
			
		case msg:=<-self.ExtComs.UpdateElevInside:
			self.ExtComs.SendMsg<-msg
			fmt.Println("\t InternalCommunication:recieved inside order update. sending to IP:", msg.To)

		case msg:=<-self.ExtComs.FailedTcpMsg:
			if msg.Type=="ADD_ORDER"{
				self.ExtComs.AddOrder<-msg.Order
				fmt.Println("\t InternalCommunicasion: Recieved faild to send ADD_ORDER msg. Taking the order self")
			}
					
		default:
			time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}
	}	
}




