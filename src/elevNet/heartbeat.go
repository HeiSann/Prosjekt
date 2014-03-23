package elevNet
import( "time"
		"elevTypes"
	   )
	   
const HEARTBEAT_TIMEOUT_MILLI = 70
const SLEEP_TIME = 20


func (elevNet *ElevNet_s) RefreshNetwork(){
	elevHeartbeatTimes:=make(map[string]time.Time)
	go elevNet.intComs.startHeartbeatTimer()
	go elevNet.broadCastHeartbeat()
	
	for{
		select{
		case msg := <-elevNet.ExtComs.HeartbeatMsg:
			elevNet.updateHeartbeatTime(elevHeartbeatTimes,msg) 
						
		case <-elevNet.intComs.timerOut:
			go elevNet.intComs.performTimeControl(elevHeartbeatTimes)
						
		case deadIp := <-elevNet.intComs.deadHeartbeater:
			elevNet.deleteHeartbeater(elevHeartbeatTimes,deadIp)
		default:
			time.Sleep(time.Millisecond*SLEEPTIME)
					
		}
	}
}


func addHeartbeater(heartbeatMap map[string]time.Time,ip string){
	_, inMap :=heartbeatMap[ip]
		
	if !inMap{
		heartbeatMap[ip]=time.Now()
	}
}


func (self *ElevNet_s) updateHeartbeatTime(heartbeatMap map[string]time.Time, msg elevTypes.Message){ 
	heartbeatIP :=msg.From
	_, inMap := heartbeatMap[heartbeatIP]
	if !inMap{
		self.intComs.connectToElev<-heartbeatIP
		self.ExtComs.NewElev<-heartbeatIP
	}
	
	limitStamp:=time.Now().Add(time.Millisecond*HEARTBEAT_TIMEOUT_MILLI)
	heartbeatMap[msg.From]=limitStamp
}


func (toRefresh *InternalChan_s)performTimeControl(heartbeatMap map[string]time.Time){
	currentTime :=time.Now()
	for ip,heartbeattime := range heartbeatMap{
		if currentTime.After(heartbeattime){			
			toRefresh.deadHeartbeater<-ip	
		}
	}
}


func (self *ElevNet_s)deleteHeartbeater(heartbeatMap map[string]time.Time, ip string){
	delete(heartbeatMap,ip)
	self.intComs.deadElev<-ip
	self.ExtComs.DeadElev<-ip	
}


func (toNet *ElevNet_s) broadCastHeartbeat(){
	myIp:=GetMyIP()
	destIp:=GetBroadcastIP(myIp)
	msg:=ConstructHeartbeat(destIp,myIp)
	for{	
		toNet.ExtComs.SendBcast<-msg
		time.Sleep(time.Millisecond*SLEEP_TIME)
	}
}


func (ToRefresh *InternalChan_s) startHeartbeatTimer(){
	for{
		time.Sleep(time.Millisecond*HEARTBEAT_TIMEOUT_MILLI)
		ToRefresh.timerOut<-true
	}
}


func ConstructHeartbeat(ipTo string, ipFrom string)elevTypes.Message{
	msg:=elevTypes.Message{}
	msg.To =ipTo
	msg.From=ipFrom
	msg.Type ="HEARTBEAT"
	return msg
}

