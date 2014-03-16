package comsManager
import( "elevTypes"
			)
  

type ComsManager_s struct{
	Ip string
	ExtComs elevTypes.ComsManager_ExtComs_s
	intComs InternalChan_s
	
}


    
type InternalChan_s struct{
	auctionWinner 	chan string
	newCostMsg 		chan elevTypes.Message
	toAuction 		chan elevTypes.Message
    auctionDone 	chan string
    costMsg 		chan elevTypes.Message
    needCost		chan elevTypes.Order_t
	
}


func InternalChannelsInit()InternalChan_s{
	intChans:=InternalChan_s{}
	intChans.auctionWinner	= make(chan string)
	intChans.newCostMsg 	= make(chan elevTypes.Message)
	intChans.toAuction 		= make(chan elevTypes.Message)
	intChans.auctionDone 	= make(chan string)
	intChans.costMsg 		= make(chan elevTypes.Message)
	intChans.needCost 		=make(chan elevTypes.Order_t)
	
	
	return intChans
}

func ExternalChannelsInit(net elevTypes.Net_ExtComs_s)elevTypes.ComsManager_ExtComs_s{
	extChans:=elevTypes.ComsManager_ExtComs_s{}
	//communication to network
	extChans.RecvMsg=net.RecvMsg
	extChans.PingMsg=net.PingMsg
	extChans.SendMsg=net.SendMsg
	extChans.SendMsgToAll=net.SendMsgToAll
	extChans.DeadElev = net.DeadElev
	//communication to order
	extChans.RequestCost = make(chan elevTypes.Order_t)
	extChans.RecvCost = make(chan int)
	extChans.AuctionOrder = make(chan elevTypes.Order_t,20)
	extChans.AddOrder = make(chan elevTypes.Order_t)
	extChans.SendOrderUpdate = make(chan elevTypes.Order_t)
	extChans.RecvOrderUpdate = make(chan elevTypes.Message)
	extChans.AuctionDeadElev =make(chan string)
	extChans.CheckNewElev = make(chan elevTypes.Message)
	extChans.UpdateElevInside =make(chan elevTypes.Message,20)
	
	return extChans

}

func Init(ip string, net elevTypes.Net_ExtComs_s)ComsManager_s{
 	comsMan := ComsManager_s{}	
	comsMan.Ip =ip	
	comsMan.ExtComs=ExternalChannelsInit(net)
	comsMan.intComs=InternalChannelsInit()
	
	go comsMan.RecieveMessageFromNet()
	go comsMan.InternalCommunication()
	go comsMan.manageAuction()
	
	return comsMan
}
