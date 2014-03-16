package elevOrders

import(
    "fmt"
	"time"
	"math"
	"sync"
	"elevTypes"
)

type Orders_s struct{
    MY_IP           string
	queues			map[string][elevTypes.N_FLOORS][elevTypes.N_DIR]bool
	emg				bool
	ExtComs			elevTypes.Orders_ExtComs_s
	mutex           sync.Mutex
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
			self.update_queue(msg.Order, msg.Payload)
			//send ACK?

		case order:= <-self.ExtComs.RequestScoreChan:
		    fmt.Println("			order.orderHandler: recieved on RequestScoreChan, order: ",order) 
			elevPos:= self.get_elev_pos()
			score := getScore(order, elevPos, self.queues[self.MY_IP]) 
			self.ExtComs.RespondScoreChan <- score
			
		case order:=<-self.ExtComs.AddOrder:
		    fmt.Println("			order.orderHandler: recieved on AddOrder, order: ",order) 
		    self.update_queue(order, self.MY_IP)
		    if self.isQueueEmpty(self.MY_IP){
		        self.ExtComs.NewOrdersChan <- order
		    }
		    
		case deadElev:= <-self.ExtComs.AuctionDeadElev:
		    self.handle_dead_elev(deadElev)
		
		case msg:= <-self.ExtComs.CheckNewElev:
		    queue := self.queues[msg.To]
		    for floor:=0 ; floor < elevTypes.N_FLOORS-1 ; floor++{
		        if queue[floor][elevTypes.NONE]{
		            msg.Order = elevTypes.Order_t{floor, elevTypes.NONE, true}
		            self.ExtComs.UpdateElevInside <- msg
		        }
		    } 
		    
		/* from FSM */
		case elevPos:= <-self.ExtComs.ExecdOrderChan:
			fmt.Println("			orders.orderHandler: got execdorder at Pos: ", elevPos)			
			order := getOrderAtPos(elevPos, self.queues[self.MY_IP])
			order.Active = false
			fmt.Println("			orders.orderHandler: found execdOrder: ", order)
			self.update_queue(order, self.MY_IP)
			fmt.Println("			orders.orderHandler: updating queue success! trying to send: self.ExtComs.SendOrderUpdate <- order. ChanID: ", self.ExtComs.SendOrderUpdate)
			self.ExtComs.SendOrderUpdate <- order
			fmt.Println("			orders.orderHandler: orderUpdate sendt! trying to check nextOrder")
			nextOrder:= get_next_order(self.queues[self.MY_IP], order)
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
			next_order:= get_next_order(queue, order)
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
			self.update_queue(msg.Order, msg.Payload)
	
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
				self.update_queue(order, self.MY_IP)
			} else{
			self.ExtComs.AuctionOrder <- order
			}
		}
		time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
	}
}

func (self *Orders_s)update_queue(order elevTypes.Order_t, IP string){
	fmt.Println("			orders.updating_queue: ", order)
	wasEmpty := self.isQueueEmpty(IP)	
	switch(order.Active){
		case true:
		    queue := self.queues[IP]
			queue[order.Floor][order.Direction] = true
			self.queues[IP] = queue
			self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, true}
			fmt.Println("			orders.updating_queue: sendt light in SetLightChan: ", elevTypes.Light_t{order.Floor, order.Direction, true})
		case false:
		    if IP == self.MY_IP{ 
				self.delete_order(order, IP)

		        next_order := get_next_order(self.queues[IP], order)
				//check for double-order executions
				fmt.Println("			orders.updating_queue: order was: ", order)
				fmt.Println("			orders.updating_queue: next_order is: ", next_order)
			 
			    if (next_order.Floor == order.Floor ){
					fmt.Println("			orders.updating_queue: Double exec!")
			        also_execd := next_order
			        self.ExtComs.SendOrderUpdate <- also_execd
			        self.delete_order(also_execd, IP)
			    }
	        }else{
	            self.delete_order(order, IP)
	        }
	}
	if wasEmpty && IP == self.MY_IP{
		fmt.Println("			orders.update_queue: sending order to fsm on NewOrderChan!")
		self.ExtComs.NewOrdersChan <- order
	}
	fmt.Println("           queues are now: ", self.queues)
}

func (self *Orders_s)get_elev_pos() elevTypes.ElevPos_t{
    pos := elevTypes.ElevPos_t{}
    self.ExtComs.ElevPosRequest <- pos
    pos =<-self.ExtComs.ElevPosRequest
    return pos
}

func (self *Orders_s)handle_dead_elev(deadElev string){
    queue := self.queues[deadElev]
    for floor:=0; floor< elevTypes.N_FLOORS; floor++{
	    for dir:=0; dir< elevTypes.N_DIR; dir++{
		    if	queue[floor][dir] == true{
		        switch dir{
		            case 0: 
		                self.ExtComs.AuctionOrder <- elevTypes.Order_t{floor,elevTypes.UP,true}
		            case 1:
		                self.ExtComs.AuctionOrder <- elevTypes.Order_t{floor,elevTypes.DOWN,true}
		            default:
		                fmt.Println("			orders.handle_dead_elev: unknown dir!")
		        }
		        queue[floor][dir] = false
		    }
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

func (self *Orders_s)delete_order(order elevTypes.Order_t, IP string){
    queue := self.queues[IP]
	queue[order.Floor][elevTypes.NONE] = false
	queue[order.Floor][order.Direction] = false
	self.queues[IP] = queue
	
	self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false} 
	self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, false}
	fmt.Println("			delete_order: deleted ", order)
	fmt.Println("			delete_order: deleted ", elevTypes.Order_t{order.Floor, elevTypes.NONE, false})
}

func (self *Orders_s)delete_all_orders_on_floor(order elevTypes.Order_t, IP string){
    queue := self.queues[IP]
   	if order.Floor == 0 || order.Floor == elevTypes.N_FLOORS-1{
   		self.delete_order(order, IP)
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

func next_order_above(this_floor int, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
    orderOut:= false
    orderUp:= false
    for floor := this_floor; floor < elevTypes.N_FLOORS; floor++{
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

func next_order_below(this_floor int, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
        orderOut    := false
        orderDown   := false
        for floor := this_floor; floor >= 0; floor--{
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

func get_next_order(queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool,order elevTypes.Order_t) elevTypes.Order_t{
    fmt.Println("			order.get_next_order: with order: ", order)
	fmt.Println("			order.get_next_order: queue is now: ", queue)
    switch(order.Direction){
        case elevTypes.UP:
            order_up_above := next_order_above(order.Floor, queue)
            if order_up_above.Active { 
                fmt.Println("			order.get_next_order: returning order_up_above: ", order_up_above)
                return order_up_above }
            
            order_down := next_order_below(elevTypes.N_FLOORS-1, queue)
            if order_down.Active { 
                fmt.Println("			order.get_next_order: returning order_down: ", order_down)
                return order_down }
            
            order_up_below := next_order_above(0, queue)
            if order_up_below.Active { 
                fmt.Println("			order.get_next_order: returning order_up_below: ", order_up_below)
                return order_up_below }
            
            /*  no orders left  */
            return elevTypes.Order_t{}
            
        case elevTypes.DOWN:
            order_down_below := next_order_below(order.Floor, queue)
            if order_down_below.Active { 
                fmt.Println("			order.get_next_order: returning order_down_below: ", order_down_below)
                return order_down_below }
            
            order_up_below := next_order_above(0, queue)
            if order_up_below.Active{ 
                fmt.Println("			order.get_next_order: returning order_up_below: ", order_up_below)
                return order_up_below }
            
            order_down_above := next_order_below(elevTypes.N_FLOORS-1, queue)
            if order_down_above.Active{ 
                fmt.Println("			order.get_next_order: returning order_down_above: ", order_down_above)
                return order_down_above}
            /*  no orders left  */
            return elevTypes.Order_t{}

		case elevTypes.NONE:
			order_up_above := next_order_above(order.Floor, queue)
            if order_up_above.Active { 
                fmt.Println("			order.get_next_order: returning order_up_above: ", order_up_above)
                return order_up_above }

			order_down_below := next_order_below(order.Floor, queue)
			if order_down_below.Active { 
                fmt.Println("			order.get_next_order: returning order_down_below: ", order_down_below)
                return order_down_below }

			order_down_above := next_order_below(elevTypes.N_FLOORS-1, queue)
            if order_down_above.Active{ 
                fmt.Println("			order.get_next_order: returning order_down_above: ", order_down_above)
                return order_down_above}

			order_up_below := next_order_above(0, queue)
            if order_up_below.Active { 
                fmt.Println("			order.get_next_order: returning order_up_below: ", order_up_below)
                return order_up_below }
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

