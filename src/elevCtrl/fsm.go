package elevCtrl

import (
	"fmt"
	"elevTypes"
	"elevDrivers"
)

type State_t int
const(
	IDLE State_t = iota
	DOORS_OPEN
	MOVING_DOWN
	MOVING_UP
	EMG
	OBST
	OBST_EMG
)

type Event_t int
const(
	start_down Event_t = iota
	start_up
	exec_order 
	timeout
	stop
	obst
)

/* 	FSM Actions */
func (self *Fsm_s)action_start_down(){
   self.Coms.motorChan <- elevTypes.DOWM
	self.state = MOVING_DOWN
	self.lastDir = elevTypes.DOWN
	fmt.Println("fsm: MOVING_DOWN\n")
}

func (self *Fsm_s)action_start_up(){
	self.Coms.motorChan <- elevTypes.UP
	self.state = MOVING_UP
	self.lastDir = elevTypes.UP
	fmt.Println("fsm: MOVING_UP\n")
}

func (self *Fsm_s)action_exec(){
	self.Coms.doorOpenChan <- true
	elevDrivers.SetLight(elev.lastFloor, elev.lastDir)    //TODO: fix with channels
	//start_timer()	
	//order_executed()
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_halt_n_exec(){
	self.Coms.motorChan <- elevTypes.NONE
	self.Coms.doorOpenChan <- true
	elevDrivers.SetLight(elev.lastFloor, elev.lastDir)    //TODO: fix with channels
	//start_timer()	
	//order_executed()
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_done(){
	self.Coms.doorOpenChan <- false
	// stop_timer()
	self.lastDir = elev.get_nearest_order()
	self.state = IDLE
	fmt.Println("fsm: IDLE\n")
	self.ready()                                          //TODO: fix 

func (self *Fsm_s)action_next(){
    self.handle_new_order()
}

func (self *Fsm_s)action_stop(){
	self.Coms.motorChan <- elevTypes.NONE
	self.state = EMG
}

func (self *Fsm_s)action_pause(){
	self.state = OBST
}

func action_dummy(){
	fmt.Println("fsm: dummy!\n")
}

/* Finite State Machine */
func (elev *Fsm_s)fsm_init(){
	elev.fsm_table = [][]func(){
/*STATES:	  \	EVENTS:	//start_down			   //start_up              //exec_order			//timeout			   //ready
/*IDLE       */  []func(){elev.action_start_down, elev.action_start_up,elev.action_exec,	      action_dummy,       elev.action_next},
/*DOORS_OPEN */  []func(){elev.action_start_down, elev.action_start_up,elev.action_exec,	      elev.action_done,   action_dummy},  
/*MOVING_UP  */  []func(){action_dummy, 		      action_dummy,        elev.action_halt_n_exec,action_dummy,       action_dummy},
/*MOVING_DOWN*/  []func(){action_dummy, 		      action_dummy,			elev.action_halt_n_exec,action_dummy,       action_dummy},
/*EMG_STOP   */  []func(){action_dummy, 		      action_dummy,			action_dummy,           action_dummy,       action_dummy},  
/*OBST       */  []func(){action_dummy, 		      action_dummy,			action_dummy,           action_dummy,       action_dummy}, 
	}
}

/* FSM help functions */
func (self *Fsm_s)fsm_update(event Event_t){
	self.fsm_table[elev.state][event]()
}

func (self *Fsm_s)should_stop(floor int) bool{     
	//TODO: communicate with orders and check if current floor got pending orders in correct direction    
	stop := true	
	return stop
}

func (self *Fsm_s)get_nearest_order() elevTypes.Direction_t{
	//TODO: communicate with orders and get next direction
	return elevTypes.UP
}

func (elev *Fsm_s)handle_new_order(){
    if elev.state == IDLE{
	    switch(elev.get_nearest_order()){
	    case elevType.UP:
	       elev.fsm_update(start_up)
	    case elevType.DOWN:
	       elev.fsm_update(start_down)
	    case elevType.NONE:
	       elev.state = IDLE
	       fmt.Println("fsm: IDLE")
	    default:
	       fmt.Println("fsm: ERROR, undefined elev.lastDir in execute_next_order")
	    }
	}else{
	    fmt.Println("fsm: new order registered, will be executed when I'm ready")
   }
}

func (elev *Fsm_s)fsm_generate_n_handle_events(){
	for{
	   select{
	   case <- elev.stopButtonChan:
		   elev.fsm_update(stop)
	   case <- elev.obsChan:
		   elev.fsm_update(obst)
	   case floor:=<- elev.floorChan:
		   if floor != -1{
			   elev.lastFloor = floor
			   //set floor_light
			   if elev.should_stop(floor){
				   elev.fsm_update(exec_order)
			   }
		   }
	   case <- elev.newOrder:
		   elev.handle_new_order()
       case <-elev.timer:
	       elev.fsm_update(timeout)	
	   case <-elev.ready:
	       elev.fsm_update(ready)
	   }
	}
}
