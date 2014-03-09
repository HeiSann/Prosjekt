package comsManager

import ("elevTypes"
		"time"
		)


const SELECT_SLEEP_TIME = 1
const AUCTION_DURATION = 30
		

func getMyCost()int{ //do this need to be in order maybe
    dummy:=0
    return dummy
}


func (fromOrder *ComsManager_s)manageAuction(){
	isAuction := false
	//currentOrder:="none"
	for{
		select{
		case order:=<-fromOrder.ExtComs.NewExtOrder: //from order module
		
			if isAuction{
				fromOrder.ExtComs.WaitAuction<-true //send to order so that order waits with sending the new auction
			}else{
				go fromOrder.intComs.auction(order) 
				isAuction =true	
			}			
		case cost:=<-fromOrder.intComs.costMsg:
			if isAuction{ //and right orderMsg
				fromOrder.intComs.newCostMsg<-cost
			}//else just throw away old/irrelevant cost msg
		case winner:=<-fromOrder.intComs.auctionDone:
			HandleAuctionWinner(winner)
			isAuction=false	
		default:
			time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}
	}		
}

func (intChans *InternalChan_s)auction(order elevTypes.Message){
    limit:=time.Now().Add(AUCTION_DURATION)
    //cost:=getMyCost()
	//winner:="my ip";
	for{
	    currentTime:=time.Now()
	    if currentTime.After(limit){
	        break
	    }
	    select{
		/*case msg:=<-intChans.newCostMsg:
		    if msg.(Payload)<cost{
		        cost=msg.payload
		        winner=msg.From	
		   	}*/ //payload=int, trouble with message type		
		default:
		    time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}//end select
      }//end for
}

func HandleAuctionWinner(winner elevTypes.Message){ //needs to know winner IP and order(if self winner, just send order directly to order module). Sends TCP to winner, and waits for ack. If no ack recieved, take the order. 
	//msg:=constructWinnerMsg(winner.From)
	//toMsgSender<-msg
}	


func constructWinnerMsg(ToIpadr string)elevTypes.Message{
    msg:=elevTypes.Message{}  
    return msg
}




