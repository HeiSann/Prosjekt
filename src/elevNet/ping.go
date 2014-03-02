package elevNet
import( "time"
        "message"
        "fmt"

       )
const PING_TIMEOUT_MILLI = 50
const SLEEP_TIME = 10
const TEST_IP = "129.241.187.255"
const MY_IP = "129.241.187.152"

//new channels fix this, only for oversikt



func (elevNet *ElevNet_s) RefreshNetwork(){
	elevPingTimes:=make(map[string]time.Time)
	go elevNet.intComs.pingTimer()
    for{
        select{
        
        case newip := <-elevNet.intComs.newPinger:
            newPinger(elevPingTimes, newip)
            fmt.Println("woho new elevator friend")
            
        case msg := <-elevNet.ExtComs.PingMsg:
			updatePingTime(elevPingTimes,msg) 
			
        case <-elevNet.intComs.timerOut:
			performTimeControl(elevPingTimes)
			
			
        case deadIp := <-elevNet.intComs.deadElev:
            deletePinger(elevPingTimes,deadIp)
            
        default:
            time.Sleep(time.Millisecond*SLEEP_TIME) //change. Needs to sleep less then a second
            elevNet.ExtComs.BroadCastPing()
            
        }//end select
    }//end for
}

func newPinger(pingMap map[string]time.Time,ip string){
	_, inMap :=pingMap[ip]
		
	if !inMap{
		pingMap[ip]=time.Now()
		fmt.Println("new pinger")
	}
}

func updatePingTime(pingMap map[string]time.Time, msg message.Message){
	
	_, inMap := pingMap[msg.From]
	if inMap{
		limitStamp:=time.Now().Add(20000000)
		pingMap[msg.From]=limitStamp
	}else{
		//handle the situation when an elevator tries to ping and is not stored in the map(not a tcp connection)
	}
}

func performTimeControl(pingMap map[string]time.Time){
	
	currentTime :=time.Now()
	for _,pingtime := range pingMap{
		if currentTime.After(pingtime){
			fmt.Println("oh no, my friend died")
	    	
	    }
    }
}

func deletePinger(pingMap map[string]time.Time, ip string){
	delete(pingMap,ip)
}

func (toNet *ExternalChan_s) BroadCastPing(){
	msg:=message.ConstructPing(TEST_IP,MY_IP)
	toNet.SendBcast<-msg
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

            
//hva hvis error når man sender melding over nettverk. Kanskje en kanal som sender den ikke sendte meldingen tilbake til comsManager som sender den dit den kom fra??

