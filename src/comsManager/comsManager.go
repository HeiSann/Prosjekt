package comsManager
import( "elevTypes"
			)
  

type ComsManager_s struct{
	ExtComs elevTypes.ComsManager_ExtComs_s
	intComs InternalChan_s
}


    
type InternalChan_s struct{
	dymmy chan int
}


func InternalChannelsInit()InternalChan_s{
	intChans:=InternalChan_s{}
	intChans.auctionWinner=make(chan string)
	return intChans
}

func ExternalChannelsInit(net elevTypes.Net_ExtComs_s)elevTypes.ComsManager_ExtComs_s{
	extChans:=elevTypes.ComsManager_ExtComs_s{}
	//communication to network
	extChans.RecvMsg=net.RecvMsg
	extChans.PingMsg=net.PingMsg
	//communication to order
	
	return extChans

}

func Init(ip string, net elevTypes.Net_ExtComs_s)ComsManager_s{

	comsMan := ComsManager_s{}	
	comsMan.ExtComs=ExternalChannelsInit(net)
	comsMan.intComs=InternalChannelsInit()
	return comsMan
	
}
