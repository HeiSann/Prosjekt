package elevFSM

import (
	"time"
	"elevTypes"
)

func Init(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s)Fsm_s{

	fsm := Fsm_s{}
	fsm.initFsmTable()
	fsm.initIntComs()
	fsm.initExtComs(driver, orders)

	fsm.start()
	
	//start the fsm routunes
	go fsm.update()
	go fsm.generateEvents()
	go fsm.answerPosRequests()
	
	return fsm
}	


/* Finite State Machine initializations */
func (fsm *Fsm_s)initFsmTable(){
	fsm.table = [][]func(){
/*STATES:	  \	EVENTS:	//NewOrder			//FloorReached			//Exec  				//TimerOut		//Obst				//EmgPressed
/*IDLE	   */  []func(){fsm.action_start,	action_dummy,			fsm.action_execSame,	action_dummy,	fsm.action_pause,	fsm.action_stop},
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
	self.intComs.floorChan			= make(chan int)
}

func(self *Fsm_s)initExtComs(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s){	
	self.ExtComs.FloorChan 			= driver.SensorChan
	self.ExtComs.StopButtonChan		= driver.StopButtonChan
	self.ExtComs.ObsChan			= driver.ObsChan	
	self.ExtComs.MotorChan			= driver.MotorChan
	self.ExtComs.DoorOpenChan		= driver.DoorOpenChan
	self.ExtComs.SetLightChan		= driver.SetLightChan
	self.ExtComs.SetFloorIndChan	= driver.SetFloorIndChan
	self.ExtComs.ElevPosRequest	 = orders.ElevPosRequest
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
	//stop motor
	self.ExtComs.MotorChan <- elevTypes.NONE
	//update fsm vars
	self.state = IDLE
	self.lastFloor = floor
	self.lastDir = elevTypes.NONE
}


func (self *Fsm_s)update(){
	var event Event_t
	for{
		event =<- self.intComs.eventChan
		self.table[self.state][event]()
	}
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
					//fsm.lastFloor = floor  //TODO: FIX!
				   	go func(){ fsm.intComs.floorChan <- floor}()
					fsm.intComs.eventChan <- FLOOR_REACHED
			   }
			   
			case order:=<- fsm.ExtComs.NewOrdersChan:
				
				go func() {fsm.intComs.eventChan <- NEW_ORDER}()
				
				go func() {fsm.intComs.newOrderChan <- order}()
				
			case execOrder:= <- fsm.ExtComs.ExecResponseChan:
				if execOrder{
					fsm.intComs.eventChan <- EXEC_ORDER
				}
				
			case <-fsm.intComs.timeoutChan:
				go func() {fsm.intComs.eventChan <- TIMEOUT}()	
				
			default:
				time.Sleep(time.Millisecond*elevTypes.SELECT_SLEEP_MS/2)	
		}
	}
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
	self.ExtComs.MotorChan <- elevTypes.DOWN
	self.state = MOVING_DOWN
	self.lastDir = elevTypes.DOWN
}


func (self *Fsm_s)startUp(){
	self.ExtComs.MotorChan <- elevTypes.UP
	self.state = MOVING_UP
	self.lastDir = elevTypes.UP
}


func startTimer(timeOutChan chan bool, timeInterval time.Duration){
	time.Sleep(time.Second*timeInterval)
	timeOutChan <- true 
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
