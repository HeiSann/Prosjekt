package comsManager

import ("time"
		"elevTypes"
		"fmt"
		)


func (comsMan *ComsManager_s)RecieveMessageFromNet(){
	for{
		msg:=<-comsMan.ExtComs.RecvMsg
		
		switch msg.Type{
				
		  case "HEARTBEAT":
		  		comsMan.ExtComs.HeartbeatMsg<-msg

		  case "COST":
		  		comsMan.intComs.costMsg<-msg
				
		  case "NEED_COST":
		  		cost :=comsMan.getMyCost(msg.Order)
		  		costMsg:=constructCostMsg(comsMan.Ip, msg.From, msg.Order, cost)
		  		comsMan.ExtComs.SendMsg<-costMsg	  		

		  case "ADD_ORDER":
		  		comsMan.ExtComs.AddOrder<-msg.Order

		  case "UPDATE_BACKUP":
				comsMan.ExtComs.RecvOrderUpdate<-msg

		default:
			fmt.Println("not able to read msg header")
	   }
	}
}


func (self *ComsManager_s)ManageCommunicationFromNetAndOrder(){ 
	for{
		select{
		case order:=<- self.ExtComs.SendOrderUpdate:
			msg:=constructUpdateMsg(self.Ip, order, self.Ip)
			self.ExtComs.SendMsgToAll<-msg
			
		case order:=<-self.intComs.needCost:
			needCostMsg:=constructNeedCostMsg(self.Ip, order)
			self.ExtComs.SendMsgToAll<-needCostMsg	
		
		case deadIp:=<-self.ExtComs.DeadElev:
			self.ExtComs.AuctionDeadElev<-deadIp		

		case newIp :=<-self.ExtComs.NewElev:
			newElevUpdate:=constructNewOrderMsg(newIp, self.Ip, elevTypes.Order_t{})
			self.ExtComs.CheckNewElev<-newElevUpdate
			
		case msg:=<-self.ExtComs.UpdateElevInside:
			self.ExtComs.SendMsg<-msg

		case msg:=<-self.ExtComs.FailedTcpMsg:
			if msg.Type=="ADD_ORDER"{
				self.ExtComs.AddOrder<-msg.Order
			}
					
		default:
			time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}
	}	
}




