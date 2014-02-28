package elevNet
import( "time"
        "message"
        "strconv"

       )
const PING_TIMEOUT_NANO = 20000
const SLEEP_TIME = 100

//new channels fix this, only for oversikt
var timerOut chan bool
var newip chan string
var pingMsg chan message.Message
var deadElev chan string
var newPing chan string



func refreshNetwork(){
	elevPingTimes:=make(map[string]int64)
    for{
        select{
        
        case newip := <-newPing:
            newPinger(elevPingTimes, newip)
            
        case msg := <-pingMsg:
			updatePingTime(elevPingTimes,msg) 
			                
        case <-timerOut:
			performTimeControl(elevPingTimes)
			
        case deadIp := <-deadElev:
            deletePinger(elevPingTimes,deadIp)
            
        default:
            time.Sleep(time.Millisecond*SLEEP_TIME) //change. Needs to sleep less then a second
            BroadCastPing()
        }//end select
    }//end for
}

func newPinger(pingMap map[string]int64,ip string){
	_, inMap :=pingMap[ip]
	if !inMap{
		pingMap[ip]=0.0
	}
}

func updatePingTime(pingMap map[string]int64, msg message.Message){
	_, inMap := pingMap[msg.From]
	if inMap{
		newtime,_:=strconv.ParseInt(msg.Payload,10,64) //Converts the message payload fro sting to int64
		pingMap[msg.From]=newtime
	}else{
		//handle the situation when an elevator tries to ping and is not stored in the map(not a tcp connection)
	}
}

func performTimeControl(pingMap map[string]int64){
	timelimit:=time.Nanosecond.Nanoseconds()*PING_TIMEOUT_NANO //converts nanosecond(type duration) to int 64 nanosecond
	currentTime :=time.Now().UnixNano() //returns current time in nanosecods
	
	for ip,pingtime := range pingMap{
		if pingtime>currentTime-timelimit{
			checkIfAlive(ip)
	    }
    }
}

func deletePinger(pingMap map[string]int64, ip string){
	delete(pingMap,ip)
}

func BroadCastPing(){
	//construct Ping msg and broadcast
	//Bcast<-pingmsg
}

func checkIfAlive(ipadr string){
    //send new tcp msg to ensure that elevator is lost
    //send msg to refresh network and updateTcpCon map(on the same channel?) so that the connection is deleted and pingmap removed
}

func pingTimer(timeOut chan bool){
    for{
        time.Sleep(time.Nanosecond*PING_TIMEOUT_NANO)
        timeOut<-true
    }
}

            
//hva hvis error når man sender melding over nettverk. Kanskje en kanal som sender den ikke sendte meldingen tilbake til comsManager som sender den dit den kom fra??

