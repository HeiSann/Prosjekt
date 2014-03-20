package elevOrders

import(
    "fmt"
	"time"
	"math"
	"elevTypes"
)

type Orders_s struct{
    MY_IP           string
	queues			map[string][elevTypes.N_FLOORS][elevTypes.N_DIR]bool
	emg				bool
	ExtComs			elevTypes.Orders_ExtComs_s
}

func Init(ip string,driver elevTypes.Drivers_ExtComs_s, coms elevTypes.ComsManager_ExtComs_s) Orders_s{
	fmt.Println("			elevOrders.init()...")

    orders := Orders_s{}
	tableMap := make(map[string][elevTypes.N_FLOORS][elevTypes.N_DIR]bool)
	var extcoms = elevTypes.Orders_ExtComs_s{}

	extcoms.ButtonChan			= driver.ButtonChan
	extcoms.SetLightChan    	= driver.SetLightChan
    extcoms.ElevPosRequest		= make(chan elevTypes.ElevPos_t)
	extcoms.NewOrdersChan		= make(chan elevTypes.Order_t)
	extcoms.ExecdOrderChan  	= make(chan elevTypes.ElevPos_t)
	extcoms.ExecRequestChan 	= make(chan elevTypes.ElevPos_t)		
	extcoms.ExecResponseChan	= make(chan bool)
	extcoms.EmgTriggerdChan 	= make(chan bool)
	extcoms.AuctionOrder		= coms.AuctionOrder
	extcoms.RequestScoreChan	= coms.RequestCost
	extcoms.RespondScoreChan	= coms.RecvCost
	extcoms.AddOrder			= coms.AddOrder
	extcoms.SendOrderUpdate 	= coms.SendOrderUpdate
	extcoms.RecvOrderUpdate 	= coms.RecvOrderUpdate
	extcoms.AuctionDeadElev     = coms.AuctionDeadElev
	extcoms.CheckNewElev        = coms.CheckNewElev
	extcoms.UpdateElevInside    = coms.UpdateElevInside

	orders.MY_IP = ip
	orders.queues = tableMap
	orders.ExtComs = extcoms
	
	go orders.orderHandler()
 	fmt.Println("			orders.init: OK!")
    return orders
}

func (self *Orders_s)orderHandler(){
	for{
	poller:
		select{
		/* from ComHandler */
		case msg:= <-self.ExtComs.RecvOrderUpdate:	
		    fmt.Println("			order.orderHandler: recieved on RecvOrderUpdate, msg: ",msg) 
			self.updateQueue(msg.Order, msg.Payload)
			//send ACK?

		case order:= <-self.ExtComs.RequestScoreChan:
		    fmt.Println("			order.orderHandler: recieved on RequestScoreChan, order: ",order) 
			elevPos:= self.getElevPos()
			score := getScore(order, elevPos, self.queues[self.MY_IP]) 
			self.ExtComs.RespondScoreChan <- score
			
		case order:=<-self.ExtComs.AddOrder:
		    fmt.Println("			order.orderHandler: recieved on AddOrder, order: ",order) 
		    self.updateQueue(order, self.MY_IP)
		    if self.isQueueEmpty(self.MY_IP){
		        self.ExtComs.NewOrdersChan <- order
		    }
		    
		case deadElev:= <-self.ExtComs.AuctionDeadElev:
		    fmt.Println("			order.orderHandler: recieved on AuctionDeadElev, deadElev: ",deadElev)
		    self.handleDeadElev(deadElev)
		
		case msg:= <-self.ExtComs.CheckNewElev:
		    fmt.Println("			order.orderHandler: recieved on CheckNewElev, msg: ",msg)
		    queue := self.queues[msg.To]
		    for floor:=0 ; floor < elevTypes.N_FLOORS ; floor++{
		        if queue[floor][elevTypes.NONE]{
		            msg.Order = elevTypes.Order_t{floor, elevTypes.NONE, true}
		            self.ExtComs.UpdateElevInside <- msg
		            fmt.Println("			order.orderHandler: found inside order, sending msg: ",msg)
		        }
		    } 
		    
		/* from FSM */
		case elevPos:= <-self.ExtComs.ExecdOrderChan:
			fmt.Println("			orders.orderHandler: got execdorder at Pos: ", elevPos)			
			order := getOrderAtPos(elevPos, self.queues[self.MY_IP])
			order.Active = false
			fmt.Println("			orders.orderHandler: found execdOrder: ", order)
			self.updateQueue(order, self.MY_IP)
			fmt.Println("			orders.orderHandler: updating queue success! trying to send: self.ExtComs.SendOrderUpdate <- order. ChanID: ", self.ExtComs.SendOrderUpdate)
			self.ExtComs.SendOrderUpdate <- order
			fmt.Println("			orders.orderHandler: orderUpdate sendt! trying to check nextOrder")
			nextOrder:= getNextOrder(self.queues[self.MY_IP], order)
			fmt.Println("			orders.orderHandler: got execdOrder: ", order)
			if nextOrder.Active{
			    self.ExtComs.NewOrdersChan <- nextOrder
			    fmt.Println("			orders.orderHandler: next order sendt to fsm: ", nextOrder)
			}

		case elevPos:= <-self.ExtComs.ExecRequestChan:
			fmt.Println("			orders.orderHandler: got execRequest: ", elevPos)
			queue := self.queues[self.MY_IP]
			order := elevTypes.Order_t{elevPos.Floor, elevPos.Direction, false}
			//this_order:= getOrderAtPos(elevPos, queue) 
			next_order:= getNextOrder(queue, order)
			fmt.Println("			orders.orderHandler: elevPos: ", elevPos)
			fmt.Println("			orders.orderHandler: next_order: ", next_order)
			if elevPos.Floor == next_order.Floor{
			    fmt.Println("			orders.orderHandler: sending true on execResponse ")
				self.ExtComs.ExecResponseChan <- true
			}else{
			    fmt.Println("			orders.orderHandler: sending false on execResponse ")
			}

		case self.emg =<-self.ExtComs.EmgTriggerdChan:
			
		case msg:= <-self.ExtComs.RecvOrderUpdate:
			self.updateQueue(msg.Order, msg.Payload)
	
		/* from driver */
		case button := <-self.ExtComs.ButtonChan:
			fmt.Println("			order.orderHandler: got button press!", button)
			order := elevTypes.Order_t{button.Floor, button.Dir, true}
			for ip, queue := range self.queues{
				if queue[order.Floor][order.Direction]{
					fmt.Println("			order.orderHandler: order already is handled by ", ip)
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
		time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
	}
}

func (self *Orders_s)updateQueue(order elevTypes.Order_t, IP string){
	fmt.Println("			orders.updating_queue: ", order)
	wasEmpty := self.isQueueEmpty(IP)	
	switch(order.Active){
		case true:
		    queue := self.queues[IP]
			queue[order.Floor][order.Direction] = true
			self.queues[IP] = queue
			if IP == self.MY_IP || order.Direction != elevTypes.NONE{
			    self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, true}
			    fmt.Println("			orders.updating_queue: sendt light in SetLightChan: ", elevTypes.Light_t{order.Floor, order.Direction, true})
			}
		case false:
		    if IP == self.MY_IP{ 
				self.deleteOrder(order, IP)

		        next_order := getNextOrder(self.queues[IP], order)
				//check for double-order executions
				fmt.Println("			orders.updating_queue: order was: ", order)
				fmt.Println("			orders.updating_queue: next_order is: ", next_order)
			 
			    if (next_order.Floor == order.Floor ){
					fmt.Println("			orders.updating_queue: Double exec!")
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
		fmt.Println("			orders.update_queue: sending order to fsm on NewOrderChan!")
		self.ExtComs.NewOrdersChan <- order
	}
	fmt.Println("           queues are now: ", self.queues)
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

func (self *Orders_s)isQueueEmpty(ip string) bool{
	//fmt.Println("			Checking queue...")
	queue:= self.queues[ip]
	for floor:=0; floor< elevTypes.N_FLOORS; floor++{
		for dir:=0; dir< elevTypes.N_DIR; dir++{
			if	queue[floor][dir] == true{
				//fmt.Println("			queue not empty!")
				return false
			}
		}
	}
	//fmt.Println("			queue empty!")
	return true
}	

func (self *Orders_s)deleteOrder(order elevTypes.Order_t, IP string){
    queue := self.queues[IP]
	queue[order.Floor][elevTypes.NONE] = false
	queue[order.Floor][order.Direction] = false
	self.queues[IP] = queue
	
	self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false} 
	self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, false}
	fmt.Println("			delete_order: deleted ", order)
	fmt.Println("			delete_order: deleted ", elevTypes.Order_t{order.Floor, elevTypes.NONE, false})
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
		fmt.Println("           delete_all_orders: deleted floor ", order.Floor)
   	}
   	
}

func nextOrderAbove(thisFloor int, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
    orderOut:= false
    orderUp:= false
    for floor := thisFloor; floor < elevTypes.N_FLOORS; floor++{
            orderOut = queue[floor][elevTypes.NONE]
            orderUp = queue[floor][elevTypes.UP]
            if orderOut{
                fmt.Println("			next_order_above: returning next: ", elevTypes.Order_t{floor, elevTypes.NONE, true}) 
                return elevTypes.Order_t{floor, elevTypes.NONE, true}
            }
            if orderUp{
                fmt.Println("			next_order_above returning next: ", elevTypes.Order_t{floor, elevTypes.UP, true}) 
                return elevTypes.Order_t{floor, elevTypes.UP, true}
            }
        }
    fmt.Println("			next_order_above found nothing, returning empty:  ", elevTypes.Order_t{}) 
    return elevTypes.Order_t{}
}

func nextOrderBelow(thisFloor int, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
        orderOut    := false
        orderDown   := false
        for floor := thisFloor; floor >= 0; floor--{
            orderOut = queue[floor][elevTypes.NONE]
            orderDown = queue[floor][elevTypes.DOWN]
            if orderOut{
                fmt.Println("			next_order_below: returning next: ", elevTypes.Order_t{floor, elevTypes.NONE, true}) 
                return elevTypes.Order_t{floor, elevTypes.NONE, true}
            }
            if orderDown{
                fmt.Println("			next_order_below: returning next: ", elevTypes.Order_t{floor, elevTypes.DOWN, true}) 
                return elevTypes.Order_t{floor, elevTypes.DOWN, true}
            }
        }
    fmt.Println("			next_order_below found nothing, returning empty:  ", elevTypes.Order_t{}) 
    return elevTypes.Order_t{}
}

func getNextOrder(queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool,order elevTypes.Order_t) elevTypes.Order_t{
    fmt.Println("			order.get_next_order: with order: ", order)
	fmt.Println("			order.get_next_order: queue is now: ", queue)
    switch(order.Direction){
        case elevTypes.UP:
            orderUpAbove := nextOrderAbove(order.Floor, queue)
            if orderUpAbove.Active { 
                fmt.Println("			order.get_next_order: returning orderUpAbove: ", orderUpAbove)
                return orderUpAbove }
            
            orderDown := nextOrderBelow(elevTypes.N_FLOORS-1, queue)
            if orderDown.Active { 
                fmt.Println("			order.get_next_order: returning orderDown: ", orderDown)
                return orderDown }
            
            orderUpBelow := nextOrderAbove(0, queue)
            if orderUpBelow.Active { 
                fmt.Println("			order.get_next_order: returning orderUpBelow: ", orderUpBelow)
                return orderUpBelow }
            
            /*  no orders left  */
            return elevTypes.Order_t{}
            
        case elevTypes.DOWN:
            orderDown_below := nextOrderBelow(order.Floor, queue)
            if orderDown_below.Active { 
                fmt.Println("			order.get_next_order: returning orderDown_below: ", orderDown_below)
                return orderDown_below }
            
            orderUpBelow := nextOrderAbove(0, queue)
            if orderUpBelow.Active{ 
                fmt.Println("			order.get_next_order: returning orderUpBelow: ", orderUpBelow)
                return orderUpBelow }
            
            orderDownAbove := nextOrderBelow(elevTypes.N_FLOORS-1, queue)
            if orderDownAbove.Active{ 
                fmt.Println("			order.get_next_order: returning orderDownAbove: ", orderDownAbove)
                return orderDownAbove}
            /*  no orders left  */
            return elevTypes.Order_t{}

		case elevTypes.NONE:
			orderUpAbove := nextOrderAbove(order.Floor, queue)
            if orderUpAbove.Active { 
                fmt.Println("			order.get_next_order: returning orderUpAbove: ", orderUpAbove)
                return orderUpAbove }

			orderDown_below := nextOrderBelow(order.Floor, queue)
			if orderDown_below.Active { 
                fmt.Println("			order.get_next_order: returning orderDown_below: ", orderDown_below)
                return orderDown_below }

			orderDownAbove := nextOrderBelow(elevTypes.N_FLOORS-1, queue)
            if orderDownAbove.Active{ 
                fmt.Println("			order.get_next_order: returning orderDownAbove: ", orderDownAbove)
                return orderDownAbove}

			orderUpBelow := nextOrderAbove(0, queue)
            if orderUpBelow.Active { 
                fmt.Println("			order.get_next_order: returning orderUpBelow: ", orderUpBelow)
                return orderUpBelow }
			return elevTypes.Order_t{}
        default:
            fmt.Println("			order.get_next_order: dir on request = NONE. this when first order is same floor")
            return order
    }
}  

func countOrders(queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool) int{
	count := 0
	for floor:=0; floor< elevTypes.N_FLOORS; floor++{
		for dir:=0; dir< elevTypes.N_DIR; dir++{
			if	queue[floor][dir] == true{
				count ++
			}
		}
	}
	return count
}

func getScore(order elevTypes.Order_t, elev elevTypes.ElevPos_t, queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool) int{
    order_already_added := queue[order.Floor][order.Direction] || queue[order.Floor][elevTypes.NONE]
	n_order := countOrders(queue)

    //Empty queue
    if n_order == 0{
		ansFloat := float64(order.Floor) - float64(elev.Floor)
		return 0 + int(math.Abs(ansFloat))

	//Existing identical order
    }else if order_already_added{
		ansFloat := float64(order.Floor) - float64(elev.Floor)
        return elevTypes.N_FLOORS + int(math.Abs(ansFloat))
        
    //Order in same direction, infront of the elevator
    }else if (elev.Direction == order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP){
        return 2*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
    }else if (elev.Direction == order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN){
        return 2*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
        
    //Order in opposite direction infront of elevator
    }else if  (elev.Direction != order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP){
        return 5*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
    }else if (elev.Direction != order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN){
        return 5*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
        
    //Order in opposite direction behind the elevator
    }else if (elev.Direction != order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP){
        return 8*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
    }else if (elev.Direction != order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN){
        return 8*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
        
    //Order in same direction behind the elevator
    }else if (elev.Direction == order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.UP){
        return 11*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
    }else if  (elev.Direction == order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.DOWN){
        return 11*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
    }
	//this shouldn't happen
	return 255
}

func MakeDouble(original Orders_s) Orders_s{
    copy := Orders_s{}
    copy.MY_IP = original.MY_IP     
    copy.queues = original.queues       
    copy.emg = original.emg           
    copy.ExtComs = original.ExtComs     
    return copy
}
