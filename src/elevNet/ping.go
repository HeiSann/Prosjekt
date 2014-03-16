package elevNet
import( "time"
        "elevTypes"
        "fmt"

       )
const PING_TIMEOUT_MILLI = 70
const SLEEP_TIME = 20
const LIMIT = 50000000


//new channels fix this, only for oversikt



func (elevNet *ElevNet_s) RefreshNetwork(){
	elevPingTimes:=make(map[string]time.Time)
	go elevNet.intComs.pingTimer()
	go elevNet.BroadCastPing()
	
    for{
        select{
    /*
        case newip := <-elevNet.intComs.newPinger:
        	fmt.Println("got new ping ip")
            addPinger(elevPingTimes, newip)
            fmt.Println("woho new elevator friend")
	*/
        case msg := <-elevNet.ExtComs.PingMsg:
			elevNet.updatePingTime(elevPingTimes,msg) 
						
        case <-elevNet.intComs.timerOut:
			go elevNet.intComs.performTimeControl(elevPingTimes)
						
        case deadIp := <-elevNet.intComs.deadPinger:
        	fmt.Println("Refresh Newtork:ping case deadIP")
            elevNet.deletePinger(elevPingTimes,deadIp)
		default:
			time.Sleep(time.Millisecond*SLEEPTIME)
                    
        }//end select
    }//end for
}

func addPinger(pingMap map[string]time.Time,ip string){
	_, inMap :=pingMap[ip]
		
	if !inMap{
		pingMap[ip]=time.Now()
		fmt.Println("new pinger")
	}
}

func (self *ElevNet_s) updatePingTime(pingMap map[string]time.Time, msg elevTypes.Message){ //change name?
	pingIP :=msg.From
	_, inMap := pingMap[pingIP]
	if !inMap{
		self.intComs.connectToElev<-pingIP
		self.ExtComs.NewElev<-pingIP
	}
	limitStamp:=time.Now().Add(time.Millisecond*PING_TIMEOUT_MILLI)
	pingMap[msg.From]=limitStamp
}

func (toRefresh *InternalChan_s)performTimeControl(pingMap map[string]time.Time){
	
	currentTime :=time.Now()
	for ip,pingtime := range pingMap{
		if currentTime.After(pingtime){
			fmt.Println("oh no, my friend died")
			toRefresh.deadPinger<-ip	
			fmt.Println("deadip sendt")    	
	    }
    }
}

func (self *ElevNet_s)deletePinger(pingMap map[string]time.Time, ip string){
	delete(pingMap,ip)
	self.intComs.deadElev<-ip
	self.ExtComs.DeadElev<-ip	
	fmt.Println("deletePinger: notified other modules about dead elevator", self.ExtComs.DeadElev)
}

func (toNet *ElevNet_s) BroadCastPing(){
	
	myIp:=GetMyIP()
	destIp:=GetBroadcastIP(myIp)
	msg:=ConstructPing(destIp,myIp)
	for{	
		toNet.ExtComs.SendBcast<-msg
		//fmt.Println("bcast sendt")
		time.Sleep(time.Millisecond*SLEEP_TIME)
	}
	
		//construct Ping msg and broadcast denne må gjennom coms manager
	//Bcast<-pingmsg
}

func checkIfAlive(ipadr string){
    //send new tcp msg to ensure that elevator is lost
    //send msg to refresh network and updateTcpCon map(on the same channel?) so that the connection is deleted and pingmap removed
}

func (ToRefresh *InternalChan_s) pingTimer(){
    for{
        time.Sleep(time.Millisecond*PING_TIMEOUT_MILLI)
        ToRefresh.timerOut<-true
    }
}

func ConstructPing(ipTo string, ipFrom string)elevTypes.Message{
    msg:=elevTypes.Message{}
	msg.To =ipTo
	msg.From=ipFrom
	msg.Type ="PING"
	return msg
}

	
            
//hva hvis error når man sender melding over nettverk. Kanskje en kanal som sender den ikke sendte meldingen tilbake til comsManager som sender den dit den kom fra??

