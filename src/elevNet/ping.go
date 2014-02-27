package elevNet
import( "time"
        "message"
        "fmt"
        "strconv"

       )
const PING_TIMEOUT_SECONDS = 1
const SLEEP_TIME = 100

//new channels fix this, only for oversikt
var timerOut chan bool
var newip chan string
var pingMsg chan message.Message
var deadElev chan string
var newPing chan string



func refreshNetwork(pingmap map[string]int64){
    for{
        select{
        case newip := <-newPing:
            pingmap[newip] = 0
        case msg := <-pingMsg:
            _, ok := pingmap[msg.From]
            if ok{
                pingmap[msg.From],_ =strconv.ParseInt(msg.Payload,10,64)//ParseInt converts string to int64 with base 10
            }else{
                fmt.Println("ip address not registered")
                //handle this. For example send connect to me msg over tcp 
            }                
        case <-timerOut:
            for ip,t := range pingmap{
                if t>time.Now().Unix()-int64(time.Second*PING_TIMEOUT_SECONDS){
                    checkIfAlive(ip)
                }
            }
        case deadIp := <-deadElev:
            delete(pingmap,deadIp)
        default:
            time.Sleep(time.Millisecond*SLEEP_TIME) //change. Needs to sleep less then a second
            //ElevNetChan.SendMsg<-Ping
        }//end select
    }//end for
}

func checkIfAlive(ipadr string){
    //send new tcp msg to ensure that elevator is lost
    //send msg to refresh network and updateTcpCon map(on the same channel?) so that the connection is deleted and pingmap removed
}

func pingTimer(timeOut chan bool){
    for{
        time.Sleep(time.Second*PING_TIMEOUT_SECONDS)
        timeOut<-true
    }
}
            
