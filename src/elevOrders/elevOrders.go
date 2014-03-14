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
   
	tableMap := make(map[string][elevTypes.N_FLOORS][elevTypes.N_DIR]bool)

	var extcoms = elevTypes.Orders_ExtComs_s{}

	extcoms.ButtonChan	= driver.ButtonChan
	extcoms.SetLightChan = driver.SetLightChan

    extcoms.ElevPosRequest      = make(chan elevTypes.ElevPos_t)
	extcoms.NewOrdersChan		= make(chan elevTypes.Order_t)
	extcoms.ExecdOrderChan  	= make(chan elevTypes.Order_t)
	extcoms.ExecRequestChan  	= make(chan elevTypes.Order_t)		
	extcoms.ExecResponseChan	= make(chan bool)
	extcoms.EmgTriggerdChan  	= make(chan bool)
	
	extcoms.AuctionOrder		= coms.AuctionOrder
	extcoms.RequestScoreChan	= coms.RequestCost
	extcoms.RespondScoreChan	= coms.RecvCost
	extcoms.AddOrder 			= coms.AddOrder
	extcoms.SendOrderUpdate 	= coms.SendOrderUpdate
	extcoms.RecvOrderUpdate 	= coms.RecvOrderUpdate

	orders := Orders_s{ip, tableMap, false, extcoms}

	go orders.orderHandler()

 	fmt.Println("			orders.init: OK!")
   return orders
}

func (self *Orders_s)orderHandler(){
	for{
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
		    
		/* from FSM */
		case order:= <-self.ExtComs.ExecdOrderChan:
			fmt.Println("			orders.orderHandler: got execdOrder: ", order)
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

		case order:= <-self.ExtComs.ExecRequestChan:
			fmt.Println("			orders.orderHandler: got execRequest: ", order)
			queue := self.queues[self.MY_IP]
			shouldExec := (doesExist(order, queue) || is_last_order_in_dir(order, queue)) 
			if shouldExec{
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
				if doesExist(order, queue){
					fmt.Println("			order.orderHandler: order already is handled by ", ip)
					break
				}
			}
			if order.Direction == elevTypes.NONE{
				self.update_queue(order, self.MY_IP)
			} else{
			fmt.Println("			order.orderHandler: Oh how exciting!! order is NEW! sending to auction! ")
			self.ExtComs.AuctionOrder <- order
			}
		}
		time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
	}
}

func doesExist(order elevTypes.Order_t, queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool) bool{
    switch(order.Direction){
        case elevTypes.UP:
            return queue[order.Floor][elevTypes.UP] || queue[order.Floor][elevTypes.NONE]
        case elevTypes.DOWN:
            return queue[order.Floor][elevTypes.DOWN] || queue[order.Floor][elevTypes.NONE]
        case elevTypes.NONE:
            fmt.Println("			order.doesExist: order.dir = NONE; this probably shouldn't happen?")
            return true
        default:
            return queue[order.Floor][elevTypes.UP] || queue[order.Floor][elevTypes.NONE]|| queue[order.Floor][elevTypes.DOWN]
    }
}

func is_last_order_in_dir(order elevTypes.Order_t, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) bool{
    fmt.Println("			order.should_turn: checking for order: ",order)
    switch order.Direction{
        case elevTypes.UP:
            fmt.Println("			order.should_turn: UP ")
            if order.Floor == elevTypes.N_FLOORS-1{
                fmt.Println("			order.should_turn: at top floor! returning true ")
                return true
            }
            for floor:= order.Floor+1; floor < elevTypes.N_FLOORS; floor ++{
                if queue[floor][elevTypes.UP] || queue[floor][elevTypes.DOWN] ||queue[floor][elevTypes.NONE]{
                    fmt.Println("			order.should_turn: found order above, returning false ",order.Floor)
                    return false
                }
            }
            fmt.Println("			order.should_turn: returning true by semidef ")
            return true
        case elevTypes.DOWN:
            fmt.Println("			order.should_turn: DOWN ")
            if order.Floor == 0 {return true}
            for floor:= order.Floor-1; floor >= 0; floor --{
                if queue[floor][elevTypes.UP] || queue[floor][elevTypes.DOWN] ||queue[floor][elevTypes.NONE]{
                    fmt.Println("			order.should_turn: found order below, returning false ",order)
                    return false
                }
            }
            return true
        
        default:
            fmt.Println("			orders.should_turn: order.dir == NONE, this probably shouldnt happen?")
            return true
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
		        next_order := get_next_order(self.queues[IP], order)
			    if next_order.Floor == order.Floor{
			        also_execd := getReverseOrder(order)
			        self.ExtComs.SendOrderUpdate <- also_execd
			        
			        self.delete_all_orders_on_floor(order, IP)
			     }else{	
			        self.delete_order(order, IP)
			    }
	        }else{
	            self.delete_order(order, IP)
	        }
	}
	
	//fmt.Println("			queue value set OK!")
	if wasEmpty && IP == self.MY_IP{
		fmt.Println("			orders.update_queue: sending order to fsm on NewOrderChan!")
		self.ExtComs.NewOrdersChan <- order
	}
	fmt.Println("           queues are now: ", self.queues)
}

func getReverseOrder(order elevTypes.Order_t) elevTypes.Order_t{
    switch order.Direction{
        case elevTypes.UP:
            return elevTypes.Order_t{order.Floor, elevTypes.DOWN, order.Active}
        case elevTypes.DOWN:
            return elevTypes.Order_t{order.Floor, elevTypes.UP, order.Active}
        default:
            fmt.Println("			orders.updating_queue: order.Dir = NONE, probably shouldn't happen?")
            return elevTypes.Order_t{order.Floor, elevTypes.NONE,  order.Active}
    }
}

func (self *Orders_s)get_elev_pos() elevTypes.ElevPos_t{
    pos := elevTypes.ElevPos_t{}
    self.ExtComs.ElevPosRequest <- pos
    pos =<-self.ExtComs.ElevPosRequest
    return pos
}


func (self *Orders_s)isQueueEmpty(ip string) bool{
	fmt.Println("			Checking queue...")
	queue:= self.queues[ip]
	for floor:=0; floor< elevTypes.N_FLOORS; floor++{
		for dir:=0; dir< elevTypes.N_DIR; dir++{
			if	queue[floor][dir] == true{
				fmt.Println("			queue not empty!")
				return false
			}
		}
	}
	fmt.Println("			queue empty!")
	return true
}	

func (self *Orders_s)delete_order(order elevTypes.Order_t, IP string){
    queue := self.queues[IP]
	queue[order.Floor][elevTypes.NONE] = false
	queue[order.Floor][order.Direction] = false
	self.queues[IP] = queue
	
	self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false} 
	self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, false}
}

func (self *Orders_s)delete_all_orders_on_floor(order elevTypes.Order_t, IP string){
    queue := self.queues[IP]
   	switch order.Floor{
   	    case 0:
   	        queue[order.Floor][elevTypes.NONE] = false
   	        queue[order.Floor][elevTypes.UP] = false
   	        self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false}
   	        self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.UP, false}
   	    case elevTypes.N_FLOORS-1:
   	        queue[order.Floor][elevTypes.NONE] = false
   	        queue[order.Floor][elevTypes.DOWN] = false
   	        self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false}
   	        self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.DOWN, false}
   	    default:
   	        queue[order.Floor][elevTypes.NONE] = false
   	        queue[order.Floor][elevTypes.UP] = false
   	        queue[order.Floor][elevTypes.DOWN] = false
   	        self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false}
   	        self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.UP, false}
   	        self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.DOWN, false}
   	}
   	self.queues[IP] = queue
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
    fmt.Println("			order.get_next_order: with dir: ", order.Direction)
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
    order_already_added := doesExist(order, queue) 
	n_order := countOrders(queue)

	//if elev is at end floor, is has to turn:
	/*	
	if elev.Floor == 0 || elev.Floor == elevTypes.N_FLOORS-1{
		elev.Direction = elevTypes.NONE
	}
	*/

    //Empty queue
    if elev.Direction == elevTypes.NONE{
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

