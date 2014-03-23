package elevOrders

import(
	"time"
	"elevTypes"
	"fmt"
)



type Orders_s struct{
	MY_IP		   string
	queues			map[string][elevTypes.N_FLOORS][elevTypes.N_DIR]bool
	emg			bool
	ExtComs			elevTypes.Orders_ExtComs_s
}

func Init(ip string, driver elevTypes.Drivers_ExtComs_s, coms elevTypes.ComsManager_ExtComs_s) Orders_s{

	orders := Orders_s{}
	tableMap := make(map[string][elevTypes.N_FLOORS][elevTypes.N_DIR]bool)
	
	extcoms := elevTypes.Orders_ExtComs_s{}	
	//Channels from Driver
	extcoms.ButtonChan			= driver.ButtonChan
	extcoms.SetLightChan		= driver.SetLightChan	
	//channels from comsManager
	extcoms.AuctionOrder		= coms.AuctionOrder
	extcoms.RequestScoreChan	= coms.RequestCost
	extcoms.RespondScoreChan	= coms.RecvCost
	extcoms.AddOrder			= coms.AddOrder
	extcoms.SendOrderUpdate 	= coms.SendOrderUpdate
	extcoms.RecvOrderUpdate 	= coms.RecvOrderUpdate
	extcoms.AuctionDeadElev		= coms.AuctionDeadElev
	extcoms.CheckNewElev		= coms.CheckNewElev
	extcoms.UpdateElevInside	= coms.UpdateElevInside	
	//Channels to FSM
	extcoms.ElevPosRequest		= make(chan elevTypes.ElevPos_t)
	extcoms.NewOrdersChan		= make(chan elevTypes.Order_t)
	extcoms.ExecdOrderChan  	= make(chan elevTypes.ElevPos_t)
	extcoms.ExecRequestChan 	= make(chan elevTypes.ElevPos_t)		
	extcoms.ExecResponseChan	= make(chan bool)
	extcoms.EmgTriggerdChan 	= make(chan bool)	

	orders.MY_IP = ip
	orders.queues = tableMap
	orders.ExtComs = extcoms
	
	go orders.orderHandler()
	return orders
}

func (self *Orders_s)orderHandler(){
	for{
	poller:
		select{
		/* from ComHandler */
		case msg:= <-self.ExtComs.RecvOrderUpdate:	 
			self.updateQueue(msg.Order, msg.Payload)

		case order:= <-self.ExtComs.RequestScoreChan:
			elevPos:= self.getElevPos()
			score := orderPlanning_getScore(order, elevPos, self.queues[self.MY_IP]) 
			self.ExtComs.RespondScoreChan <- score
			
		case order:=<-self.ExtComs.AddOrder: 
			self.updateQueue(order, self.MY_IP)
			
		case deadElev:= <-self.ExtComs.AuctionDeadElev:
			self.handleDeadElev(deadElev)
		
		case msg:= <-self.ExtComs.CheckNewElev:
			// fill out empty msg with old inside-order and send back
			newElev:=msg.To
			queue := self.queues[newElev]
			for floor:=0 ; floor < elevTypes.N_FLOORS ; floor++{
				if queue[floor][elevTypes.NONE]{
					msg.Order = elevTypes.Order_t{floor, elevTypes.NONE, true}
					self.ExtComs.UpdateElevInside <- msg
				}
			} 
			
		/* from FSM */
		case elevPos:= <-self.ExtComs.ExecdOrderChan:		
			order := getOrderAtPos(elevPos, self.queues[self.MY_IP])
			order.Active = false
			self.updateQueue(order, self.MY_IP)
			self.ExtComs.SendOrderUpdate <- order
			nextOrder:= orderPlanning_getNextOrder(self.queues[self.MY_IP], order)
			// if orders pending, send next order to elevFSM
			if nextOrder.Active{
				self.ExtComs.NewOrdersChan <- nextOrder
			}

		case elevPos:= <-self.ExtComs.ExecRequestChan:
			queue := self.queues[self.MY_IP]
			order := elevTypes.Order_t{elevPos.Floor, elevPos.Direction, false} 
			next_order:= orderPlanning_getNextOrder(queue, order)
			if elevPos.Floor == next_order.Floor{
				self.ExtComs.ExecResponseChan <- true
			}

		case self.emg =<-self.ExtComs.EmgTriggerdChan:
			
		case msg:= <-self.ExtComs.RecvOrderUpdate:
			self.updateQueue(msg.Order, msg.Payload)
	
		/* from driver */
		case button := <-self.ExtComs.ButtonChan:
			order := elevTypes.Order_t{button.Floor, button.Dir, true}
			for _, queue := range self.queues{
				if queue[order.Floor][order.Direction]{
					//order already exists in queue, don't update
					break poller
				}
			}
			if order.Direction == elevTypes.NONE{
				self.updateQueue(order, self.MY_IP)
				self.ExtComs.SendOrderUpdate <- order
			} else{
			self.ExtComs.AuctionOrder <- order
			}
		}
		time.Sleep(time.Millisecond*elevTypes.SELECT_SLEEP_MS)
	}
}


func (self *Orders_s)updateQueue(order elevTypes.Order_t, IP string){
	wasEmpty := self.isQueueEmpty(IP)	
	switch(order.Active){
		case true:
			queue := self.queues[IP]
			queue[order.Floor][order.Direction] = true
			self.queues[IP] = queue
			//only handle orders that are outside-orders or our own orders
			if IP == self.MY_IP || order.Direction != elevTypes.NONE{
				self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, true}
			}
		case false:
			if IP == self.MY_IP{ 
				self.deleteOrder(order, IP)

				next_order := orderPlanning_getNextOrder(self.queues[IP], order)
				//check for double-order executions
			 
				if (next_order.Floor == order.Floor ){
					
					also_execd := next_order
					also_execd.Active = false
					self.ExtComs.SendOrderUpdate <- also_execd
					self.deleteOrder(also_execd, IP)
				}
			}else{
				self.deleteOrder(order, IP)
			}
	}
	if wasEmpty && IP == self.MY_IP{
		//Send new order to elevFSM
		self.ExtComs.NewOrdersChan <- order
	}
}


func (self *Orders_s)isQueueEmpty(ip string) bool{
	
	queue:= self.queues[ip]
	for floor:=0; floor< elevTypes.N_FLOORS; floor++{
		for dir:=0; dir< elevTypes.N_DIR; dir++{
			if	queue[floor][dir] == true{
				return false
			}
		}
	}
	return true
}	


func (self *Orders_s)getElevPos() elevTypes.ElevPos_t{
	pos := elevTypes.ElevPos_t{}
	self.ExtComs.ElevPosRequest <- pos
	pos =<-self.ExtComs.ElevPosRequest
	return pos
}


func (self *Orders_s)handleDeadElev(deadElev string){
	queue := self.queues[deadElev]
	for floor:=0; floor< elevTypes.N_FLOORS; floor++{
		if	queue[floor][elevTypes.UP] == true{
			self.ExtComs.AuctionOrder <- elevTypes.Order_t{floor,elevTypes.UP,true}
			queue[floor][elevTypes.UP] = false
			
		}else if queue[floor][elevTypes.DOWN] == true{
			self.ExtComs.AuctionOrder <- elevTypes.Order_t{floor,elevTypes.DOWN,true}
			queue[floor][elevTypes.DOWN] = false
		}
	}
	self.queues[deadElev] = queue
}


func getOrderAtPos(elevPos elevTypes.ElevPos_t, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
	if queue[elevPos.Floor][elevPos.Direction]{
		return elevTypes.Order_t{elevPos.Floor, elevPos.Direction, true}
	}else if queue[elevPos.Floor][elevTypes.NONE]{
		return elevTypes.Order_t{elevPos.Floor, elevTypes.NONE, true}
	}else if queue[elevPos.Floor][elevTypes.UP]{
		return elevTypes.Order_t{elevPos.Floor, elevTypes.UP, true}
	}else if queue[elevPos.Floor][elevTypes.DOWN]{
		return elevTypes.Order_t{elevPos.Floor, elevTypes.DOWN, true}	
	}else{
		return 	elevTypes.Order_t{}
	}
}


func (self *Orders_s)deleteOrder(order elevTypes.Order_t, IP string){
	queue := self.queues[IP]
	queue[order.Floor][elevTypes.NONE] = false
	queue[order.Floor][order.Direction] = false
	self.queues[IP] = queue
	
	self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false} 
	self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, false}
}


func (self *Orders_s)deleteAllOrdersOnFloor(order elevTypes.Order_t, IP string){
	queue := self.queues[IP]
   	if order.Floor == 0 || order.Floor == elevTypes.N_FLOORS-1{
   		self.deleteOrder(order, IP)
	}else{
		queue = self.queues[IP]
		queue[order.Floor][elevTypes.NONE] = false
		queue[order.Floor][elevTypes.UP] = false
		queue[order.Floor][elevTypes.DOWN] = false
		self.queues[IP] = queue
		self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false}
		self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.UP, false}
		self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.DOWN, false}
   	}
   	
}


func MakeDouble(original Orders_s) Orders_s{
	copy := Orders_s{}
	copy.MY_IP = original.MY_IP	 
	copy.queues = original.queues	   
	copy.emg = original.emg		   
	copy.ExtComs = original.ExtComs	 
	return copy
}
