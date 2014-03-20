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
	MOVING_UP
	MOVING_DOWN
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
	execOrderChan	chan elevTypes.Order_t
}

type Fsm_s struct{
    table		[][]func()
	state 		State_t
	lastDir		elevTypes.Direction_t
	lastFloor   int 
	intComs		intComs_s
	ExtComs     elevTypes.Fsm_ExtComs_s
}

/* Finite State Machine Actions */
func (self *Fsm_s)action_start(){
    //fmt.Println("				fsm.action_start") 
	order := <- self.intComs.newOrderChan
	fmt.Println("				fsm.action_start: got order on intComs.newOrderChan: ", order)
	curr_floor:= self.lastFloor	
	switch {
		case order.Floor == curr_floor:
			//fmt.Println("				fsm.action_start: trying to send exec_order on int.eventChan")
			self.intComs.eventChan <- EXEC_ORDER
			//fmt.Println("				fsm.action_start: exec_order sendt int.eventChan")
		case order.Floor < curr_floor:
			self.startDown()
		case order.Floor > curr_floor:
			self.startUp()
	}
}

func (self *Fsm_s)action_checkOrder(){
    //fmt.Println("				fsm.action_checkOrder")
	
	self.ExtComs.SetFloorIndChan <- self.lastFloor
	fmt.Println("				fsm.action_checkOrder: floorIndSignal sendt")
	
	current := elevTypes.ElevPos_t{self.lastFloor, self.lastDir, true}
	self.ExtComs.ExecRequestChan <- current
	fmt.Println("				fsm.action_checkOrder: sendt to orders on Ext.Coms.ExecRecuest: ", current)
}

func (self *Fsm_s)action_execSame(){
    //fmt.Println("				fsm.action_execSame")
	self.ExtComs.DoorOpenChan <- true
	go startTimer(self.intComs.timeoutChan, elevTypes.DOOR_OPEN_TIME)
	self.lastDir = elevTypes.NONE
	self.state = DOORS_OPEN 
	fmt.Println("				fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_exec(){
    fmt.Println("				action_exec")
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.ExtComs.DoorOpenChan <- true
	go startTimer(self.intComs.timeoutChan, elevTypes.DOOR_OPEN_TIME)
	self.state = DOORS_OPEN 
	fmt.Println("				fsm: DOORS_OPEN\n")
}

func (self *Fsm_s)action_done(){
    //fmt.Println("				fsm.action_done")
	self.ExtComs.DoorOpenChan <- false
	self.state = IDLE
	fmt.Println("				fsm: IDLE")
	self.ExtComs.ExecdOrderChan <- elevTypes.ElevPos_t{self.lastFloor, self.lastDir, false}
	fmt.Println("				fsm.action_done: sendt on ExComs.ExecdOrderChan: ", elevTypes.ElevPos_t{self.lastFloor, self.lastDir, false})
}   

func (self *Fsm_s)action_stop(){
	self.ExtComs.MotorChan <- elevTypes.NONE
	self.state = EMG_STOP
}

func (self *Fsm_s)action_pause(){
	self.state = OBSTRUCTED
}

func action_dummy(){
	fmt.Println("				fsm: dummy!\n")
}

/* Finite State Machine initializations */
func (fsm *Fsm_s)initFsmTable(){
	fsm.table = [][]func(){
/*STATES:	  \	EVENTS:	//NewOrder			//FloorReached        	//Exec  				//TimerOut		//Obst				//EmgPressed
/*IDLE       */  []func(){fsm.action_start,	action_dummy,			fsm.action_execSame,	action_dummy,	fsm.action_pause,	fsm.action_stop},
/*DOORS_OPEN */  []func(){action_dummy,		action_dummy,			action_dummy,			fsm.action_done,fsm.action_pause,	fsm.action_stop},  
/*MOVING_UP  */  []func(){action_dummy,		fsm.action_checkOrder,	fsm.action_exec,		action_dummy,	fsm.action_pause,	fsm.action_stop},
/*MOVING_DOWN*/  []func(){action_dummy,		fsm.action_checkOrder,	fsm.action_exec,		action_dummy,	fsm.action_pause,	fsm.action_stop},
/*EMG_STOP	 */  []func(){action_dummy,		action_dummy,			action_dummy,			action_dummy,	fsm.action_pause,	fsm.action_stop},  
/*OBST		 */  []func(){action_dummy,		action_dummy,			action_dummy,			action_dummy,	action_dummy,		fsm.action_stop}, 
/*OBST+EMG	 */	 []func(){action_dummy,		action_dummy,			action_dummy,			action_dummy,	action_dummy,		fsm.action_stop}, 
	}
}

func(self *Fsm_s)initIntComs(){
	self.intComs.eventChan			= make(chan Event_t, 2)
	self.intComs.timeoutChan		= make(chan bool)
	self.intComs.newOrderChan		= make(chan elevTypes.Order_t)
}

func(self *Fsm_s)initExtComs(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s){	
	self.ExtComs.ButtonChan 		= driver.ButtonChan
	self.ExtComs.FloorChan 			= driver.SensorChan
	self.ExtComs.StopButtonChan	    = driver.StopButtonChan
	self.ExtComs.ObsChan			= driver.ObsChan	
	self.ExtComs.MotorChan			= driver.MotorChan
	self.ExtComs.DoorOpenChan		= driver.DoorOpenChan
	self.ExtComs.SetLightChan		= driver.SetLightChan
	self.ExtComs.SetFloorIndChan	= driver.SetFloorIndChan
	self.ExtComs.ElevPosRequest     = orders.ElevPosRequest
	self.ExtComs.NewOrdersChan		= orders.NewOrdersChan
	self.ExtComs.ExecdOrderChan 	= orders.ExecdOrderChan
	self.ExtComs.ExecRequestChan 	= orders.ExecRequestChan 
	self.ExtComs.ExecResponseChan   = orders.ExecResponseChan
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
	fmt.Println("				Found floor")
	//stop motor
	self.ExtComs.MotorChan <- elevTypes.NONE
	fmt.Println("				 floor sendt")
	//update fsm vars
	self.state = IDLE
	self.lastFloor = floor
	self.lastDir = elevTypes.NONE
}

func Init(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s)Fsm_s{
   fmt.Println("				fsm.init()...")

	fsm := Fsm_s{}
	fsm.initFsmTable()
	fsm.initIntComs()
	fsm.initExtComs(driver, orders)

	fsm.start()
	
	//start the fsm routunes
	go fsm.update()
	go fsm.generateEvents()
	go fsm.answerPosRequests()
	
	fmt.Println("				fsm.state is: ", fsm.state)
	fmt.Println("				fsm.lastFloor is: ", fsm.lastFloor)
	fmt.Println("				fsm.Init: OK!")
	return fsm
}
    

/* FSM help functions */

func (self *Fsm_s)answerPosRequests(){
    pos:= elevTypes.ElevPos_t{}
    for{
        pos= <-self.ExtComs.ElevPosRequest
        pos.Floor = self.lastFloor
        pos.Direction = self.lastDir
        self.ExtComs.ElevPosRequest <- pos
    }
}
func (self *Fsm_s)startDown(){
	fmt.Println("				fsm.start_down")
	self.ExtComs.MotorChan <- elevTypes.DOWN
	self.state = MOVING_DOWN
	self.lastDir = elevTypes.DOWN
	fmt.Println("				fsm: MOVING_DOWN\n")
}

func (self *Fsm_s)startUp(){
    fmt.Println("				fsm.start_up")
	self.ExtComs.MotorChan <- elevTypes.UP
	self.state = MOVING_UP
	self.lastDir = elevTypes.UP
	fmt.Println("				fsm: MOVING_UP\n")
}

func (self *Fsm_s)update(){
	var event Event_t
	for{
		event =<- self.intComs.eventChan
		fmt.Println("				fmt.update: updating fsm[state][event]:", self.state, " ", event)
		self.table[self.state][event]()
	}
}

func startTimer(timeOutChan chan bool, timeInterval time.Duration){
    time.Sleep(time.Second*timeInterval)
    timeOutChan <- true 
}

func (fsm *Fsm_s)generateEvents(){	
	for{
        select{
            case <-fsm.ExtComs.StopButtonChan:
			    fsm.intComs.eventChan <- EMG
			    
            case <-fsm.ExtComs.ObsChan:
			    fsm.intComs.eventChan <- OBSTRUCTION
			    
            case floor := <-fsm.ExtComs.FloorChan: 
                if floor != -1 && floor != fsm.lastFloor{
                    fsm.lastFloor = floor  //TODO: FIX!
				    fmt.Println("				fsm.generate_events: reached new floor! lastFloor is now: ", floor) 
				    fsm.intComs.eventChan <- FLOOR_REACHED
		       }
		       
            case order:=<- fsm.ExtComs.NewOrdersChan:
			    fmt.Println("				fsm: got new order on NewOrdersChan")
			    
                go func() {fsm.intComs.eventChan <- NEW_ORDER}()
                fmt.Println("				fsm: sendt order on internal eventchan")
                
			    go func() {fsm.intComs.newOrderChan <- order}()
			    fmt.Println("				fsm: sendt order on internal newOrderChan")
			    
		    case execOrder:= <- fsm.ExtComs.ExecResponseChan:
		        fmt.Println("				fsm.generate_events: got ExecResponse: ", execOrder)
			    if execOrder{
			        fmt.Println("				fsm.generate_events: sending EXEC_ORDER on intComs.eventChan")
				    fsm.intComs.eventChan <- EXEC_ORDER
			    }
			    
            case <-fsm.intComs.timeoutChan:
                fmt.Println("				fsm.generate_eventes: TIMEOUT")
			    go func() {fsm.intComs.eventChan <- TIMEOUT}()	
			    
		    default:
			    time.Sleep(time.Millisecond*elevTypes.SLOW_DOWM_MUTHA_FUKKA/2)	
		}
	}
}

func MakeDouble(original Fsm_s) Fsm_s{
	copy := Fsm_s{}
    copy.table = original.table		
	copy.state = original.state	
	copy.lastDir = original.lastDir	
	copy.lastFloor = original.lastFloor  
	copy.intComs = original.intComs
	copy.ExtComs = original.ExtComs
	return copy
}