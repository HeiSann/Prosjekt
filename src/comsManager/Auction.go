package comsManager

import ("elevTypes"
		"time"
		"strconv"
		"fmt"
		)

const AUCTION_DURATION = 70
		

func (self *ComsManager_s)manageAuction(){
	for{
		order:=<-self.ExtComs.AuctionOrder
		fmt.Println("\t manageAuction: recieved up/down order from order module")
		go self.startAuction(order)
		self.intComs.needCost<-order
		fmt.Println("\t manageAuctions: sendt needcost on internal channel to comsMan")
WhileAuction:
		for{
			select{			
			case costMsg:=<-self.intComs.costMsg:
				fmt.Println("\tmanageAuction: recieved costMsg from net")
				if (costMsg.Order.Direction==order.Direction) && (costMsg.Order.Floor==order.Floor){
					fmt.Println("\t manageAuction: right order, trying to send it to auction on channel",self.intComs.newCostMsg)
					self.intComs.newCostMsg<-costMsg
					fmt.Println("\t ManageAuction: recieved cost from other elev",costMsg.Payload)
				}
				fmt.Println("\tmanageAuction:wrong order")
			case winner:=<-self.intComs.auctionDone:
				fmt.Println("\t manageAuction recieved winner, started handle", self.intComs.auctionDone)
				self.HandleAuctionWinner(winner, order)
				fmt.Println("\t manageAuction: started handle winner")
				break WhileAuction

			default:
				time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
			}
		}
		fmt.Println("\t manageAuction: Handle auctionWinner ok, broke out of inner for loop")
	}		
}


func (coms *ComsManager_s)startAuction(order elevTypes.Order_t){
	fmt.Println("\t auction:started auction of order", order)
	limit:=time.Now().Add(time.Millisecond*AUCTION_DURATION)
	
	cost:=coms.getMyCost(order)
	fmt.Println("\t auction: got own cost", cost) 
	winner:=coms.Ip
	fmt.Println("\t auction: will read on channel:",coms.intComs.newCostMsg)
	for{
		currentTime:=time.Now()
		fmt.Println("\t auction :",currentTime)
		
		select{
		case msg:=<-coms.intComs.newCostMsg:
			temp,_:=strconv.Atoi(msg.Payload)
			fmt.Println("\t auction: recieved cost", temp)
			if temp<cost{
				cost=temp
				winner=msg.From	
		   	} 		
		default:
			time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
			fmt.Println("\t auction: default")
		}
		if currentTime.After(limit){
			fmt.Println("\tauction: timeout")
			break
		}
	  }
	coms.intComs.auctionDone<-winner
	fmt.Println("\t auction: auction done. winner sendt to auction manager", coms.intComs.auctionDone)
}


func (self *ComsManager_s)HandleAuctionWinner(winner string, order elevTypes.Order_t ){  
	OrderUpdate := constructUpdateMsg(self.Ip ,order,winner)
	
	if winner==self.Ip{
		self.ExtComs.AddOrder<-order
		fmt.Println("\t HandleAuctionWinner: sendt winner=self to self",winner)
	}else if winner!=self.Ip{
		//send update to self, in case other elevators die not gettin updated
		self.ExtComs.RecvOrderUpdate<-OrderUpdate 
		//send update to winner
		msg:= constructNewOrderMsg(winner,self.Ip, order)
		self.ExtComs.SendMsg<-msg 
		fmt.Println("\t HandleAuctionWinner: sendt winner to winner:",winner)
	}	
	self.ExtComs.SendMsgToAll<-OrderUpdate	
	fmt.Println("\t HandleAuctionWinner: send update on tcp to all",winner)
}	


func (self *ComsManager_s)getMyCost(order elevTypes.Order_t)int{ 
	self.ExtComs.RequestCost<-order
	cost:=<-self.ExtComs.RecvCost
	return cost
}


