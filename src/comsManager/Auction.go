package comsManager

import ("elevTypes"
		"time"
		"strconv"
		)

const AUCTION_DURATION = 70
		

func (self *ComsManager_s)manageAuction(){
	for{
		order:=<-self.ExtComs.AuctionOrder
		go self.startAuction(order)
		self.intComs.needCost<-order
WhileAuction:
		for{
			select{			
			case costMsg:=<-self.intComs.costMsg:
				if (costMsg.Order.Direction==order.Direction) && (costMsg.Order.Floor==order.Floor){
					self.intComs.newCostMsg<-costMsg
				}
			case winner:=<-self.intComs.auctionDone:
				self.HandleAuctionWinner(winner, order)
				break WhileAuction

			default:
				time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
			}
		}
	}		
}


func (coms *ComsManager_s)startAuction(order elevTypes.Order_t){
	limit:=time.Now().Add(time.Millisecond*AUCTION_DURATION)
	
	cost:=coms.getMyCost(order)
	winner:=coms.Ip
	for{
		currentTime:=time.Now()
		
		select{
		case msg:=<-coms.intComs.newCostMsg:
			temp,_:=strconv.Atoi(msg.Payload)
			if temp<cost{
				cost=temp
				winner=msg.From	
		   	} 		
		default:
			time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}
		if currentTime.After(limit){
			break
		}
	  }
	coms.intComs.auctionDone<-winner
}


func (self *ComsManager_s)HandleAuctionWinner(winner string, order elevTypes.Order_t ){  
	OrderUpdate := constructUpdateMsg(self.Ip ,order,winner)
	
	if winner==self.Ip{
		self.ExtComs.AddOrder<-order
	}else if winner!=self.Ip{
		//send update to self, in case other elevators die not getting updated
		self.ExtComs.RecvOrderUpdate<-OrderUpdate 
		//send update to winner
		msg:= constructNewOrderMsg(winner,self.Ip, order)
		self.ExtComs.SendMsg<-msg 
	}	
	self.ExtComs.SendMsgToAll<-OrderUpdate	
}	


func (self *ComsManager_s)getMyCost(order elevTypes.Order_t)int{ 
	self.ExtComs.RequestCost<-order
	cost:=<-self.ExtComs.RecvCost
	return cost
}


