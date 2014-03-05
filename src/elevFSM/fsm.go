package elevFSM

import (
	"fmt"
	"time"
	"elevTypes"
)

/*	Declaration of states, events and fsm_structs */
type State_t int
const(
	IDLE State_t = iota
	DOORS_OPEN
	MOVING_DOWN
	MOVING_UP
	EMG_STOP
	OBSTRUCTED
	OBST_EMG
)

type Event_t int
const(
	NEW_ORDER Event_t = iota
	FLOOR_REACHED
	EXEC_ORDER 
	TIMEOUT
	OBSTRUCTION
	EMG
)

type intComs_s struct{
   eventChan		chan Event_t
	startTimerChan	chan bool
	timeoutChan		chan bool
	newOrderChan 	chan elevTypes.Order_t
}

type Fsm_s struct{
   table			[][]func()
	state 		State_t
	lastDir		elevTypes.Direction_t
	lastFloor   int 
	intComs		intComs_s
	ExtComs     elevTypes.Fsm_ExtComs_s
}

/* FSM_init */
func Init(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s)Fsm_s{
   fmt.Println("fsm.init()...")

	fsm := Fsm_s{}
	fsm.init_fsm_table()
	fsm.init_intComs()
	fsm.init_ExtComs(driver, orders)

	fsm.start()
	
	fmt.Println("fsm.Init: OK!")
	return fsm
}

/* Finite State Machine Actions */
func (self *Fsm_s)action_start(){
	order := <- self.intComs.newOrderChan
	curr_floor:= self.lastFloor	
	switch {
		case order.Floor == curr_floor:
			 self.intComs.eventChan <- EXEC_ORDER
		case order.Floor < curr_floor:
			self.start_down()
		case order.Floor > curr_floor:
			self.start_up()
	}
}

func (self *Fsm_s)action_check_order(){
	//get_floor
	floor := <- self.ExtComs.FloorChan
	//TODO: update_lights
	
	//send_stop_request order
	current := elevTypes.Order_t{floor, self.lastDir, true}
	self.ExtComs.ExecRequestChan <- current
	resp :=<- self.ExtComs.ExecResponseChan
	if resp{
		self.intComs.eventChan <- EXEC_ORDER
	}
}

func (self *Fsm_s)action_exec_same(){
	self.ExtComs.DoorOpenChan <- true
	self.ExtComs.SetLightChan <- elevTypes.Light_t{self.lastFloor, self.lastDir, false}
	self.intComs.startTimerChan <- true
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_exec(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.ExtComs.DoorOpenChan <- true
	self.ExtComs.SetLightChan <- elevTypes.Light_t{self.lastFloor, self.lastDir, false}	//turn buttonLight off
	//TODO: open door
	self.intComs.startTimerChan <- true	
	self.ExtComs.ExecdOrderChan <- elevTypes.Order_t{self.lastFloor, self.lastDir, false}
	self.state = DOORS_OPEN 
	fmt.Println("fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_done(){
	self.ExtComs.DoorOpenChan <- false
	self.state = IDLE
	fmt.Println("fsm: IDLE\n")
	self.ExtComs.ExecdOrderChan <- elevTypes.Order_t{self.lastFloor, self.lastDir, false}
}   

func (self *Fsm_s)action_stop(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.state = EMG_STOP
}

func (self *Fsm_s)action_pause(){
	self.state = OBSTRUCTED
}

func action_dummy(){
	fmt.Println("fsm: dummy!\n")
}

/* Finite State Machine initializations */
func (fsm *Fsm_s)init_fsm_table(){
	fsm.table = [][]func(){
/*STATES:	  \	EVENTS:	//NewOrder			//FloorReached        	//Exec  					//TimerOut		//Obst					//EmgPressed
/*IDLE       */  []func(){fsm.action_start,	action_dummy,				fsm.action_exec_same,action_dummy,	fsm.action_pause,		fsm.action_stop},
/*DOORS_OPEN */  []func(){action_dummy,		action_dummy,				action_dummy,			fsm.action_done,fsm.action_pause,	fsm.action_stop},  
/*MOVING_UP  */  []func(){action_dummy,		fsm.action_check_order,	fsm.action_exec,		action_dummy,	fsm.action_pause,		fsm.action_stop},
/*MOVING_DOWN*/  []func(){action_dummy,		fsm.action_check_order,	fsm.action_exec,		action_dummy,	fsm.action_pause,		fsm.action_stop},
/*EMG_STOP   */  []func(){action_dummy,		action_dummy,				action_dummy,			action_dummy,	fsm.action_pause,		fsm.action_stop},  
/*OBST       */  []func(){action_dummy,		action_dummy,				action_dummy,			action_dummy,	action_dummy,			fsm.action_stop}, 
/*OBST+EMG	 */  []func(){action_dummy,		action_dummy,				action_dummy,			action_dummy,	action_dummy,			fsm.action_stop}, 
	}
}

func(self *Fsm_s)init_intComs(){
	self.intComs.eventChan 			= make(chan Event_t)
	self.intComs.startTimerChan 	= make(chan bool)
	self.intComs.timeoutChan		= make(chan bool)
	self.intComs.newOrderChan  	= make(chan elevTypes.Order_t)
}

func(self *Fsm_s)init_ExtComs(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s){	
	self.ExtComs.ButtonChan 		= driver.ButtonChan
	self.ExtComs.FloorChan 			= driver.SensorChan
	self.ExtComs.StopButtonChan	= driver.StopButtonChan
	self.ExtComs.ObsChan				= driver.ObsChan	
	self.ExtComs.MotorChan			= driver.MotorChan
	self.ExtComs.DoorOpenChan		= driver.DoorOpenChan
	self.ExtComs.SetLightChan		= driver.SetLightChan
	self.ExtComs.SetFloorIndChan	= driver.SetFloorIndChan
	self.ExtComs.NewOrdersChan		= orders.NewOrdersChan
	self.ExtComs.ExecdOrderChan 	= orders.ExecdOrderChan
	self.ExtComs.ExecRequestChan 	= orders.ExecRequestChan 
	self.ExtComs.EmgTriggerdChan	= orders.EmgTriggerdChan
}

func (self *Fsm_s)start(){
	floor := <- self.ExtComs.FloorChan
	if floor == -1{
		self.ExtComs.MotorChan <- elevTypes.DOWN
	}
	for floor == -1{
		floor = <- self.ExtComs.FloorChan
	}	
	fmt.Println("Found floor")
	//stop motor
	self.ExtComs.MotorChan <- elevTypes.NONE
	fmt.Println(" floor sendt")
	//update fsm vars
	self.state = IDLE
	self.lastFloor = floor
	self.lastDir = elevTypes.NONE
	//start the fsm routunes
	go self.update()
	go self.generate_events()
}

/* FSM help functions */
func (self *Fsm_s)start_down(){
   self.ExtComs.MotorChan <- elevTypes.DOWN
	self.state = MOVING_DOWN
	self.lastDir = elevTypes.DOWN
	fmt.Println("fsm: MOVING_DOWN\n")
}

func (self *Fsm_s)start_up(){
	self.ExtComs.MotorChan <- elevTypes.UP
	self.state = MOVING_UP
	self.lastDir = elevTypes.UP
	fmt.Println("fsm: MOVING_UP\n")
}

func (self *Fsm_s)update(){
	var event Event_t
	for{
		event =<- self.intComs.eventChan
		self.table[self.state][event]()
		fmt.Println("fmt.update: updated fsm[state][event]:", self.state, " ", event)
		time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)
	}
}

func (fsm *Fsm_s)generate_events(){	
	for{
	   select{
	   case <- fsm.ExtComs.StopButtonChan:
			fsm.intComs.eventChan <- EMG
	   case <- fsm.ExtComs.ObsChan:
			fsm.intComs.eventChan <- OBSTRUCTION
	   case floor:=<- fsm.ExtComs.FloorChan:
			fmt.Println("...")
		   if floor != -1 && floor != fsm.lastFloor{
				fmt.Println("fsm.generate_events: reached new floor!") 
				fsm.intComs.eventChan <- FLOOR_REACHED
		   }
	   case order:=<- fsm.ExtComs.NewOrdersChan:
			fmt.Println("fsm: ah, waking up")
		   fsm.intComs.eventChan <- NEW_ORDER
			fmt.Println("fsm: brushing teeth")
			fsm.intComs.newOrderChan <- order
			fmt.Println("fsm: woke up")
		case execOrder:= <- fsm.ExtComs.ExecResponseChan:
			if execOrder{
				fsm.intComs.eventChan <- EXEC_ORDER
			}
      case <-fsm.intComs.timeoutChan:
			 fsm.intComs.eventChan <- TIMEOUT		
		default:
			time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA)	
		}
	}
}

