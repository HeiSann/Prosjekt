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

	orders := Orders_s{table,false,extcoms}

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
			self.update_queue(order)
			//ExtComs.OrderToNetChan <- order

		case order:= <-self.ExtComs.ExecRequestChan:
			shouldExec := self.doesExist(order)
			if shouldExec{
				self.ExtComs.ExecResponseChan <- true
			}

		case self.emg =<-self.ExtComs.EmgTriggerdChan:
	
		/* from driver */
		case button := <-self.ExtComs.ButtonChan:
			fmt.Println("got button press!")
			if button.Dir == elevTypes.NONE{
				order := elevTypes.Order_t{button.Floor, button.Dir, true}
				self.update_queue(order)
			}
		}
		time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
	}
}

func (self *Orders_s)doesExist(order elevTypes.Order_t) bool{
	return self.queue[order.Floor][elevTypes.UP] || self.queue[order.Floor][elevTypes.NONE]|| self.queue[order.Floor][elevTypes.NONE]
}

func (self *Orders_s)update_queue(order elevTypes.Order_t){
	fmt.Println("Updating queue...")
	wasEmpty := self.isQueueEmpty()	
	fmt.Println("setting order in queue:", order)
	self.queue[order.Floor][order.Direction] = true
	fmt.Println("queue value set OK!")
	if wasEmpty{
		fmt.Println("Waking up fsm!")
		self.ExtComs.NewOrdersChan <- order
	}
	fmt.Println("queue updated!")
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

func getScore(order elevTypes.Order_t){
}

func clear_list(){
}

