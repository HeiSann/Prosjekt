package elevOrders

import(
   "fmt"
	"time"
   "elevTypes"
)

type Orders_s struct{
   queue          [elevTypes.N_FLOORS][elevTypes.N_DIR]bool
	//netQueues		map{string}[][]bool
	emg				bool
   ExtComs			elevTypes.Orders_ExtComs_s
}

func Init(driver elevTypes.Drivers_ExtComs_s) Orders_s{
   fmt.Println("elevOrders.init()...")
   
   var table [elevTypes.N_FLOORS][elevTypes.N_DIR]bool

	var extcoms = elevTypes.Orders_ExtComs_s{}

	extcoms.ButtonChan	= driver.ButtonChan
	extcoms.SetLightChan = driver.SetLightChan

	extcoms.NewOrdersChan		= make(chan elevTypes.Order_t)
	extcoms.ExecdOrderChan  	= make(chan elevTypes.Order_t)
	extcoms.ExecRequestChan  	= make(chan elevTypes.Order_t)		
	extcoms.ExecResponseChan	= make(chan bool)
	extcoms.EmgTriggerdChan  	= make(chan bool)

	orders := Orders_s{table, false, extcoms}

	go orders.orderHandler()

 	fmt.Println("orders.init: OK!")
   return orders
}

func (self *Orders_s)orderHandler(){
	for{
		select{
		/* from ComHandler */
		case order:= <-self.ExtComs.OrderToMeChan:	
			self.update_queue(order)
			//send ACK?

		//case order:= <-self.ExtComs.RequestScoreChan:
			//score := self.getScore(order) 
			//self.ExtComs.RespondScoreChan <- score

		/* from FSM */
		case order:= <-self.ExtComs.ExecdOrderChan:
			fmt.Println("orders.orderHandler: got execdOrder: ", order)
			self.update_queue(order)
			//ExtComs.OrderToNetChan <- order
			nextOrder:= get_next_order(self.queue, order)
			if nextOrder.Status{
			    self.ExtComs.NewOrdersChan <- nextOrder
			    fmt.Println("orders.orderHandler: next order sendt to fsm: ", nextOrder)
			}

		case order:= <-self.ExtComs.ExecRequestChan:
			fmt.Println("orders.orderHandler: got execRequest: ", order)
			shouldExec := self.doesExist(order)
			if shouldExec{
			    fmt.Println("orders.orderHandler: sending true on execResponse ")
				self.ExtComs.ExecResponseChan <- true
			}else{
			    fmt.Println("orders.orderHandler: sending false on execResponse ")
			}

		case self.emg =<-self.ExtComs.EmgTriggerdChan:
	
		/* from driver */
		case button := <-self.ExtComs.ButtonChan:
			fmt.Println("order.orderHandler: got button press!", button)
			order := elevTypes.Order_t{button.Floor, button.Dir, true}
			self.update_queue(order)
		}
		time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
	}
}

func (self *Orders_s)doesExist(order elevTypes.Order_t) bool{
    switch(order.Direction){
        case elevTypes.UP:
            return self.queue[order.Floor][elevTypes.UP] || self.queue[order.Floor][elevTypes.NONE]
        case elevTypes.DOWN:
            return self.queue[order.Floor][elevTypes.DOWN] || self.queue[order.Floor][elevTypes.NONE]
        case elevTypes.NONE:
            fmt.Println("order.doesExist: order.dir = NONE; this probably shouldn't happen?")
            return true
        default:
            return self.queue[order.Floor][elevTypes.UP] || self.queue[order.Floor][elevTypes.NONE]|| self.queue[order.Floor][elevTypes.DOWN]
    }
}

func (self *Orders_s)update_queue(order elevTypes.Order_t){
	fmt.Println("orders.updating_queue: ", order)
	fmt.Println("queue was: ", self.queue)
	wasEmpty := self.isQueueEmpty()	
	if order.Status{
	    self.queue[order.Floor][order.Direction] = order.Status
	    self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, true}
	    fmt.Println("orders.updating_queue: sendt light in SetLightChan: ", elevTypes.Light_t{order.Floor, order.Direction, true})
	}else{
	    self.queue[order.Floor][elevTypes.NONE] = order.Status
	    self.queue[order.Floor][order.Direction] = order.Status
	    self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, elevTypes.NONE, false}
	    fmt.Println("orders.updating_queue: sendt light in SetLightChan: ", elevTypes.Light_t{order.Floor, elevTypes.NONE, false})
	    self.ExtComs.SetLightChan <- elevTypes.Light_t{order.Floor, order.Direction, false}
	    fmt.Println("orders.updating_queue: sendt light in SetLightChan: ", elevTypes.Light_t{order.Floor, order.Direction, false})
	}
	//fmt.Println("queue value set OK!")
	if wasEmpty{
		fmt.Println("orders.update_queue: sending order to fsm on NewOrderChan!")
		self.ExtComs.NewOrdersChan <- order
	}
	fmt.Println("queue is now: ", self.queue)
}

func (self *Orders_s)isQueueEmpty() bool{
	fmt.Println("Checking queue...")
	for floor:=0; floor< elevTypes.N_FLOORS; floor++{
		for dir:=0; dir< elevTypes.N_DIR; dir++{
			if	self.queue[floor][dir] == true{
				fmt.Println("queue not empty!")
				return false
			}
		}
	}
	fmt.Println("queue empty!")
	return true
}	

func next_order_above(this_floor int, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
    orderOut:= false
    orderUp:= false
    for floor := this_floor; floor < elevTypes.N_FLOORS; floor++{
            orderOut = queue[floor][elevTypes.NONE]
            orderUp = queue[floor][elevTypes.UP]
            if orderOut{
                fmt.Println("next_order_above: returning next: ", elevTypes.Order_t{floor, elevTypes.NONE, true}) 
                return elevTypes.Order_t{floor, elevTypes.NONE, true}
            }
            if orderUp{
                fmt.Println("next_order_above returning next: ", elevTypes.Order_t{floor, elevTypes.UP, true}) 
                return elevTypes.Order_t{floor, elevTypes.UP, true}
            }
        }
    fmt.Println("next_order_above found nothing, returning empty:  ", elevTypes.Order_t{}) 
    return elevTypes.Order_t{}
}

func next_order_below(this_floor int, queue[elevTypes.N_FLOORS][elevTypes.N_DIR]bool) elevTypes.Order_t{
        orderOut    := false
        orderDown   := false
        for floor := this_floor; floor >= 0; floor--{
            orderOut = queue[floor][elevTypes.NONE]
            orderDown = queue[floor][elevTypes.DOWN]
            if orderOut{
                fmt.Println("next_order_below: returning next: ", elevTypes.Order_t{floor, elevTypes.NONE, true}) 
                return elevTypes.Order_t{floor, elevTypes.NONE, true}
            }
            if orderDown{
                fmt.Println("next_order_below: returning next: ", elevTypes.Order_t{floor, elevTypes.DOWN, true}) 
                return elevTypes.Order_t{floor, elevTypes.DOWN, true}
            }
        }
    fmt.Println("next_order_below found nothing, returning empty:  ", elevTypes.Order_t{}) 
    return elevTypes.Order_t{}
}

func get_next_order(queue [elevTypes.N_FLOORS][elevTypes.N_DIR]bool,order elevTypes.Order_t) elevTypes.Order_t{
    fmt.Println("order.get_next_order: with dir: ", order.Direction)
    switch(order.Direction){
        case elevTypes.UP:
            order_up_above := next_order_above(order.Floor, queue)
            if order_up_above.Status { 
                fmt.Println("order.get_next_order: returning order_up_above: ", order_up_above)
                return order_up_above }
            
            order_down := next_order_below(elevTypes.N_FLOORS-1, queue)
            if order_down.Status { 
                fmt.Println("order.get_next_order: returning order_down: ", order_down)
                return order_down }
            
            order_up_below := next_order_above(0, queue)
            if order_up_below.Status { 
                fmt.Println("order.get_next_order: returning order_up_below: ", order_up_below)
                return order_up_below }
            
            /*  no orders left  */
            return elevTypes.Order_t{}
            
        case elevTypes.DOWN:
            order_down_below := next_order_below(order.Floor, queue)
            if order_down_below.Status { 
                fmt.Println("order.get_next_order: returning order_down_below: ", order_down_below)
                return order_down_below }
            
            order_up_below := next_order_above(0, queue)
            if order_up_below.Status{ 
                fmt.Println("order.get_next_order: returning order_up_below: ", order_up_below)
                return order_up_below }
            
            order_down_above := next_order_below(elevTypes.N_FLOORS-1, queue)
            if order_down_above.Status{ 
                fmt.Println("order.get_next_order: returning order_down_above: ", order_down_above)
                return order_down_above}
             
            /*  no orders left  */
            return elevTypes.Order_t{}
         default:
            fmt.Println("order.get_next_order: dir on request = NONE. this probably shouldn't happen?")
            return elevTypes.Order_t{}
    }
}  


func clear_list(){
}

func getScore(order elevTypes.Order_t, elev elevTypes.Order_t) int{
    order_already_added := does_exist(order)
    //Empty queue
    if elev.Direction == elevTypes.NONE
        return 0 + abs(order.Floor-elev.Floor);
        
    //Existing identical order
    elseif order_already_added
        return elevTypes.N_FLOORS + abs(order.Floor-elev.Floor)
        
    //Order in same direction, infront of the elevator
    elseif (elev.Direction == order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP)
        return 2*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
    elseif (elev.Direction == order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN)
        return 2*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
        
    //Order in opposite direction infront of elevator
    elseif (elev.Direction ~= order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP)
        return 5*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
    elseif (elev.Direction ~= order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN)
        return 5*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
        
    //Order in opposite direction behind the elevator
    elseif (elev.Direction ~= order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.UP)
        return 8*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
    elseif (elev.Direction ~= order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.DOWN)
        return 8*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
        
    //Order in same direction behind the elevator
    elseif (elev.Direction == order.Direction) && (order.Floor<elev.Floor) && (order.Direction==elevTypes.UP)
        return 11*elevTypes.N_FLOORS + elev.Floor-order.Floor + 2*n_order
    elseif (elev.Direction == order.Direction) && (order.Floor>elev.Floor) && (order.Direction==elevTypes.DOWN)
        return 11*elevTypes.N_FLOORS + order.Floor-elev.Floor + 2*n_order
    }
}

