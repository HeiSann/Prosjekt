package comsManager

include("elevTypes"
		"time"
		)
		
		
func auction(){
	myCost:= computeCost()
	cost:=myCost
	
	select
	case msg:=<-recieveCost:
		if msg.payload<cost{
			cost=msg.payload
	case <-auctionTimeOut:
		to comsMan
	default:
		time.Sleep(time.Millisecond*2)
			
	

}


func computeCost(){ //do this need to be in order maybe
}

func manageAuction(order)
	for{
		order<-newAuction
		b
		go auction()
		result:=<-auctionDone
		
