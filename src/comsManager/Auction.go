package comsManager

include("elevTypes"
		"time"
		)

const SELECT_SLEEP_TIME = 1
		
func auction(endTime time.time, order elevTypes.order){
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


func getMycost(){ //do this need to be in order maybe
}

func manageAuction()
	isAuction := true
	for{
		select{
		case order:<-newAuction:
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

func HandleAuctionWinner(winnerIP string){
	msg:=constructWinnerMsg(winnerIP)
	<-
}	

func constructWinnerMsg(ToIpadr string)elevTypes.Message{
}

