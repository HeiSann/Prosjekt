package comsManager

import ("elevTypes"
		"time"
		)


const SELECT_SLEEP_TIME = 1
const AUCTION_DURATION = 30
		

func (coms *ComsManager_s)getMyCost(order elevTypes.Order_t)int{ //do this need to be in order maybe
	coms.ExtComs.RequestCost<-order
	cost:=<-coms.ExtComs.RecvOrder
    return cost
}


func (coms *ComsManager_s)manageAuction(){

	for{
		order:=<-coms.ExtComs.AuctionOrder:
		go fromOrder.intComs.auction(order)
Auction:
		for{
			select{			
			case cost:=<-coms.intComs.costMsg:
				if cost.Order.Direction=order.Direction && cost.Order.Floor==order.Floor{
					coms.intComs.newCostMsg<-cost
				}//check if correct order 
			case winner:=<-coms.intComs.auctionDone:
				HandleAuctionWinner(winner, order)
				break Auction

			default:
				time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
			}
		}
	}		
}

func (intChans *InternalChan_s)auction(order elevTypes.Order_t){
    limit:=time.Now().Add(AUCTION_DURATION)
    cost:=getMyCost(order)
	winner:="my ip";
	for{
	    currentTime:=time.Now()
	    if currentTime.After(limit){
	        break
	    }
	    select{
		case msg:=<-intChans.newCostMsg:
		    if msg.Payload<cost{
		        cost=msg.payload
		        winner=msg.From	
		   	} //payload=int, trouble with message type		
		default:
		    time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}//end select
      }//end for
	intChans.auctionDone<-winner
}

func (coms *ComsManager_s)HandleAuctionWinner(winner string, order elevTypes.oder_t ){ //needs to know winner IP and order(if self winner, just send order directly to order module). Sends TCP to winner, and waits for ack. If no ack recieved, take the order. 
	if winner==ComsManager.Ip{
		coms.ExtComs.addOrder<-order
	}
	toAll= constructUpdateMsg(winner,order)
	coms.ExtComs.SendMsgToAll<-toAll	
	msg:=constructNotifyWinnerMsg(winner,order)
	coms.ExtComs.SendMsg<-msg
	
}	


func constructUpdateMsg(winner string, order elevTypes.Order)elevTypes.Message{
	msg:=elevTypes.Message{}
	msg.From=ComsMan.Ip
	msg.Type="UPDATE_BACKUP"
	msg.Payload = winner
	msg.Order = order	 
}

func constructNewOrderMsg(ToIpadr string, order elevTypes.Order)elevTypes.Message(
	msg:=elevTypes.Message{} 
	msg.To=ToIpadr
	msg.From = ComsManager.Ip
	msg.Type = "ADD_ORDER" 
   	msg.Order= order
	return msg
}
