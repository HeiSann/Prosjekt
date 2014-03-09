package comsManager
/*
import ("elevTypes"
		"time"
		)


const SELECT_SLEEP_TIME = 1
const AUCTION_DURATION = 20
		

func getMycost(){ //do this need to be in order maybe
}

func manageAuction(){
	isAuction := false
	for{
		select{
		case order:=<-newExtOrder: //from order module
			if isAuction{
				waitAuction<-true //send to order so that order waits with sending the new auction
			}else{
				go auction(order)
				isAuction =true	
			}			
		case cost:=<-newCostmsg:
			if isAuction{
				toAuction<-cost
			}//else just throw away old/irrelevant cost msg
		case winner:=<-auctionDone:
			HandleAuctionWinner(winner)
			isAuction=false	
		default:
				time.Sleep(time.Millisecons*SELECT_SLEEP_TIME)
		}
	}		
}

func auction(endTime time.Time, order elevTypes.order){
	for time.Now()<endTime{
		cost:=getMyCost()
		winner:="my ip";
	
		select{
		case msg:=<-recieveCost:
			//check if correct order, throw away if old order cost
			if msg.payload<cost{
				cost=msg.payload
				winner=msg.From			
		default:
			time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}
	auctionDone<-winner
	}
}

func HandleAuctionWinner(winner order){ //needs to know winner IP and order(if self winner, just send order directly to order module). Sends TCP to winner, and waits for ack. If no ack recieved, take the order. 
	msg:=constructWinnerMsg(winnerIP)
	toMsgSender<-msg
}	

func constructWinnerMsg(ToIpadr string)elevTypes.Message{
    msg:=elevTypes.msg{}  
    return msg
}
*/



