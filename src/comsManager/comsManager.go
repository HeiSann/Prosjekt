package comsManager
import( "elevTypes"
			)
  

type ComsManager_s struct{
	Ip string
	ExtComs elevTypes.ComsManager_ExtComs_s
	intComs InternalChan_s
	
}


    
type InternalChan_s struct{
	auctionWinner chan string
	newCostMsg chan elevTypes.Message
	toAuction chan elevTypes.Message
   auctionDone chan elevTypes.Message
   costMsg chan elevTypes.Message
	
}


func InternalChannelsInit()InternalChan_s{
	intChans:=InternalChan_s{}
	intChans.auctionWinner=make(chan string)
	intChans.newCostMsg = make(chan elevTypes.Message)
	intChans.toAuction = make(chan elevTypes.Message)
	intChans.auctionDone = make(chan elevTypes.Message)
	intChans.costMsg = make(chan elevTypes.Message)
	
	
	return intChans
}

func ExternalChannelsInit(net elevTypes.Net_ExtComs_s)elevTypes.ComsManager_ExtComs_s{
	extChans:=elevTypes.ComsManager_ExtComs_s{}
	//communication to network
	extChans.RecvMsg=net.RecvMsg
	extChans.PingMsg=net.PingMsg
	extChans.SendMsg=net.SendMsg
	//communication to order
	extChans.RequestCost chan int
	extChans.RecvCost chan int
	extChans.AuctionOrder chan elevTypes.Order_t
	extChans.AddOrder chan elevTypes.Order_t
	extChans.SendOrderUpdate chan elevTypes.Order_t
	extChans.RecvOrderUpdate chan elevTypes.Order_t
	
	return extChans

}

func Init(ip string, net elevTypes.Net_ExtComs_s)ComsManager_s{
 	Ip :=ip
	comsMan := ComsManager_s{}	
	comsMan.ExtComs=ExternalChannelsInit(net)
	comsMan.intComs=InternalChannelsInit()
	return comsMan
	
}
