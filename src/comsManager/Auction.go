package comsManager

import ("elevTypes"
		"time"
		"strconv"
		)


const SELECT_SLEEP_TIME = 1
const AUCTION_DURATION = 30
		

func (self *ComsManager_s)getMyCost(order elevTypes.Order_t)int{ //do this need to be in order maybe
	self.ExtComs.RequestCost<-order
	cost:=<-self.ExtComs.RecvCost
    return cost
}


func (self *ComsManager_s)manageAuction(){

	for{
		order:=<-self.ExtComs.AuctionOrder
		go self.auction(order)
Auction:
		for{
			select{			
			case costMsg:=<-self.intComs.costMsg:
				if (costMsg.Order.Direction==order.Direction) && (costMsg.Order.Floor==order.Floor){
					self.intComs.newCostMsg<-costMsg
				}//check if correct order 
			case winner:=<-self.intComs.auctionDone:
				self.HandleAuctionWinner(winner, order)
				break Auction

			default:
				time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
			}
		}
	}		
}

func (coms *ComsManager_s)auction(order elevTypes.Order_t){
    limit:=time.Now().Add(AUCTION_DURATION)
    cost:=coms.getMyCost(order)
	winner:="MY_IP"
	for{
	    currentTime:=time.Now()
	    if currentTime.After(limit){
	        break
	    }
	    select{
		case msg:=<-coms.intComs.newCostMsg:
			temp,_:=strconv.Atoi(msg.Payload)
		    if temp<cost{
				cost=temp
		        winner=msg.From	
		   	} //payload=int, trouble with message type		
		default:
		    time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}//end select
      }//end for
	coms.intComs.auctionDone<-winner
}

func (self *ComsManager_s)HandleAuctionWinner(winner string, order elevTypes.Order_t ){ //needs to know winner IP and order(if self winner, just send order directly to order module). Sends TCP to winner, and waits for ack. If no ack recieved, take the order. 
	if winner==self.Ip{
		self.ExtComs.AddOrder<-order
	}
	toAll := constructUpdateMsg(self.Ip ,order,winner)
	self.ExtComs.SendMsgToAll<-toAll	
	msg:= constructNewOrderMsg(winner,self.Ip, order)
	self.ExtComs.SendMsg<-msg
	
}	


func constructUpdateMsg(myIp string, order elevTypes.Order_t, actionElev string)elevTypes.Message{
	msg:=elevTypes.Message{}
	msg.From=myIp
	msg.Type="UPDATE_BACKUP"
	msg.Payload = actionElev
	msg.Order = order	 
	return msg
}

func constructNewOrderMsg(ToIpadr string, myIp string, order elevTypes.Order_t)elevTypes.Message{
	msg:=elevTypes.Message{} 
	msg.To=ToIpadr
	msg.From = myIp
	msg.Type = "ADD_ORDER" 
   	msg.Order= order
	return msg
}
