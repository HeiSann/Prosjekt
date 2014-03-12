package comsManager

import ("fmt"
		"time"
		"elevTypes"
		)




func (comsMan *ComsManager_s)RecieveMessageFromNet(){
    for{
        msg:=<-comsMan.ExtComs.RecvMsg
        
    
        switch msg.Type{
		  case "test":
				fmt.Println("tcp msg recieved")
				comsMan.TcpSenderTest(msg.From)
				
		  case "PING":
		  		comsMan.ExtComs.PingMsg<-msg

		  case "COST":
		  		comsMan.intComs.costMsg<-msg

		  case "NEED_COST":
		  		//send cost function and the order it relates to. JSON for order send?? 

		  case "ADD_ORDER":
		  		comsMan.ExtComs.AddOrder<-msg.Order

		  case "UPDATE_BACKUP":
				comsMan.ExtComs.RecvOrderUpdate<-msg

		  case "DEAD ELEVATOR":
		  //broadcast the ip. Start auctioning the dead elevators external orders

		  case "TCP ERROR":
		  	//go routine
		  	//try to send msg againg
		  	//ig msg fails n times. Send msg to the sender that the msg was lost. Take the order if the msg was an order type

        default:
            fmt.Println(msg.From)
            fmt.Println("not able to read msg header. Something went terribly wrong, oh god, i have dissapointed the other elevators, they will hate me so much. Pls ctrlC me right now I cant stand this pain any longer")
       }
    }
}

func (self *ComsManager_s)TcpSenderTest(to string){
        msg:=elevTypes.Message{}
        msg.From= "129.241.187.156"
        msg.To= to
        msg.Type="test"
        self.ExtComs.SendMsg<-msg 
}


func (self *ComsManager_s)ForwardMessageFromOrder(){
	for{
		select{
		case order:=<- self.ExtComs.SendOrderUpdate:
			msg:=constructUpdateMsg(self.Ip, order, self.Ip)
			self.ExtComs.SendMsgToAll<-msg
			time.Sleep(time.Millisecond*SELECT_SLEEP_TIME)
		}
	}	
}

