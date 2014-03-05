package elevNet
import( "time"
        "elevTypes"
        "fmt"

       )
const PING_TIMEOUT_MILLI = 50
const SLEEP_TIME = 30
const LIMIT = 50000000


//new channels fix this, only for oversikt



func (elevNet *ElevNet_s) RefreshNetwork(){
	elevPingTimes:=make(map[string]time.Time)
	go elevNet.intComs.pingTimer()
	go elevNet.ExtComs.BroadCastPing()
    for{
        select{
    /*
        case newip := <-elevNet.intComs.newPinger:
        	fmt.Println("got new ping ip")
            addPinger(elevPingTimes, newip)
            fmt.Println("woho new elevator friend")
	*/
        case msg := <-elevNet.ExtComs.PingMsg:
			elevNet.intComs.updatePingTime(elevPingTimes,msg) 
						
        case <-elevNet.intComs.timerOut:
			go elevNet.intComs.performTimeControl(elevPingTimes)
						
        case deadIp := <-elevNet.intComs.deadPinger:
        	fmt.Println("ping case deadIP")
            elevNet.intComs.deletePinger(elevPingTimes,deadIp)
                    
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

func (toTcp *InternalChan_s) updatePingTime(pingMap map[string]time.Time, msg elevTypes.Message){ //change name?
	pingIP :=msg.From
	_, inMap := pingMap[pingIP]
	if !inMap{
		toTcp.connectToElev<-pingIP
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

func (toTcp *InternalChan_s)deletePinger(pingMap map[string]time.Time, ip string){
	delete(pingMap,ip)
	toTcp.deadElev<-ip	
}

func (toNet *ExternalChan_s) BroadCastPing(){
	myIp:=GetMyIP()
	destIp:=GetBroadcastIP(myIp)
	for{	
		msg:=ConstructPing(destIp,myIp)
		toNet.SendBcast<-msg
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
    return elevTypes.Message{ipTo, ipFrom, "PING", ""}
}

	
            
//hva hvis error når man sender melding over nettverk. Kanskje en kanal som sender den ikke sendte meldingen tilbake til comsManager som sender den dit den kom fra??

