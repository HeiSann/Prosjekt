package comsManager

import ("elevTypes"
		"time"
		"strconv"
		"fmt"
		)


const SELECT_SLEEP_TIME = 2
const AUCTION_DURATION = 50
		

func (self *ComsManager_s)getMyCost(order elevTypes.Order_t)int{ 
	self.ExtComs.RequestCost<-order
	cost:=<-self.ExtComs.RecvCost
    return cost
}


func (self *ComsManager_s)manageAuction(){
	for{
		order:=<-self.ExtComs.AuctionOrder
		fmt.Println("\t manageAuction: recieved up/down order from order module")
		go self.auction(order)
		self.intComs.needCost<-order
		fmt.Println("\t manageAuctions: sendt needcost on internal channel to comsMan")
Auction:
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
				break Auction

			default:
				time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
			}
		}
		fmt.Println("\t manageAuction: Handle auctionWinner ok, broke out of inner for loop")
	}		
}

func (coms *ComsManager_s)auction(order elevTypes.Order_t){
	fmt.Println("\t auction:started auction of order", order)
    limit:=time.Now().Add(time.Millisecond*AUCTION_DURATION)
    
    cost:=coms.getMyCost(order)
    fmt.Println("\t auction: got cost", cost) //debug
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
		}//end select
		if currentTime.After(limit){
	    	fmt.Println("\tauction: timeout")
	        break
	    }
      }//end for
	coms.intComs.auctionDone<-winner
	fmt.Println("\t auction: auction done. winner sendt to auction manager", coms.intComs.auctionDone)
}

func (self *ComsManager_s)HandleAuctionWinner(winner string, order elevTypes.Order_t ){  
	if winner==self.Ip{
		self.ExtComs.AddOrder<-order
		fmt.Println("\t HandleAuctionWinner: sendt winner=self",winner)
	}
	
	
	msg:= constructNewOrderMsg(winner,self.Ip, order)
	self.ExtComs.SendMsg<-msg
	fmt.Println("\t HandleAuctionWinner: send order to winner", winner)
	
	toAll := constructUpdateMsg(self.Ip ,order,winner)
	fmt.Println("\t HandleAuctionWinner:constructed update msg, trying to send on channel:", self.ExtComs.SendMsgToAll)
	self.ExtComs.SendMsgToAll<-toAll	
	fmt.Println("\t HandleAuctionWinner: send update on tcp to all",winner)
	
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
