package elevOrders

import(
   "fmt"
   "elevTypes"
)

type Orders_s struct{
   queue          [][]bool
	netQueues		map{string}[][]bool
	emg				bool
   ExtComs			elevTypes.Orders_ExtComs_s
}

func Init() Orders_s{
   fmt.Println("elevOrders.init()...")
   
   var table [N_FLOORS][N_DIR]bool
	var extcoms = elevTypes.Orders_ExtComs_s{}

	NewOrdersChan    	make(chan elevTypes.Order_t) 
   OrderUpdatedChan	make(chan elevTypes.Order_t)
	OrderExecdChan  	make(chan elevTypes.Order_t)
	StopRequestChan  	make(chan elevTypes.Order_t)		
   EmgTriggerdChan  	make(chan bool)
 
   return Orders_s{}
}

func orderHandler(){
	for{
		select{
		/* from ComHandler */
		case order:= <-ExtComs.OrderFromNetChan:
			wasEmpty := isQueueEmpty()			
			update_queue(order)
			if wasEmpty{
				Extcoms.NewOrdersChan <- order
			}
		case order:= <-ExtComs.RequestScoreChan:
			score := getScore(order) 
			ExtComs.RespondScoreChan <- score
		/* from FSM */
		case order:= <-ExtComs.OrderExecdChan:
			update_queue(order)
			ExtComs.OrderToNetChan <- order
		case order:= <-stopRequestChan:
			shouldExec := doesExist(order)
			if shouldExec{
				ExtComs.ExecRespondChan <- true
			}
		case emg <-EmgTriggerdChan:
			if emg{
				self.emg = true
			}else{
				self.emg = false
			}
		}	
	}
}

func doesExist(elevTypes.Order_t) bool{
}

func update_queue(order){
}

func clear_list(){
}

func getScore(order elevTypes.Order_t){
}

func startAuction(){

}

