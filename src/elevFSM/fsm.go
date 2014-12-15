package elevFSM

import (
    "fmt"
    "time"
    "elevTypes"
)

func Init(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s)Fsm_s{
   fmt.Println("                fsm.init()...")

    fsm := Fsm_s{}
    fsm.initFsmTable()
    fsm.initIntComs()
    fsm.initExtComs(driver, orders)

    fsm.start()
    
    //start the fsm routunes
    go fsm.update()
    go fsm.generateEvents()
    go fsm.answerPosRequests()
    
    fmt.Println("               fsm.state is: ", fsm.state)
    fmt.Println("               fsm.lastFloor is: ", fsm.lastFloor)
    fmt.Println("               fsm.Init: OK!")
    return fsm
}   


/* Finite State Machine initializations */
func (fsm *Fsm_s)initFsmTable(){
    fsm.table = [][]func(){
/*STATES:     \ EVENTS: //NewOrder          //FloorReached          //Exec                  //TimerOut      //Obst              //EmgPressed
/*IDLE     */  []func(){fsm.action_start,   action_dummy,           fsm.action_execSame,    action_dummy,   fsm.action_pause,   fsm.action_stop},
/*DOORS_OPEN */  []func(){action_dummy,     action_dummy,           action_dummy,           fsm.action_done,fsm.action_pause,   fsm.action_stop},  
/*MOVING_UP  */  []func(){action_dummy,     fsm.action_checkOrder,  fsm.action_exec,        action_dummy,   fsm.action_pause,   fsm.action_stop},
/*MOVING_DOWN*/  []func(){action_dummy,     fsm.action_checkOrder,  fsm.action_exec,        action_dummy,   fsm.action_pause,   fsm.action_stop},
/*EMG_STOP   */  []func(){action_dummy,     action_dummy,           action_dummy,           action_dummy,   fsm.action_pause,   fsm.action_stop},  
/*OBST       */  []func(){action_dummy,     action_dummy,           action_dummy,           action_dummy,   action_dummy,       fsm.action_stop}, 
/*OBST+EMG   */  []func(){action_dummy,     action_dummy,           action_dummy,           action_dummy,   action_dummy,       fsm.action_stop}, 
    }
}

func(self *Fsm_s)initIntComs(){
    self.intComs.eventChan          = make(chan Event_t, 2)
    self.intComs.timeoutChan        = make(chan bool)
    self.intComs.newOrderChan       = make(chan elevTypes.Order_t)
    self.intComs.floorChan          = make(chan int)
}

func(self *Fsm_s)initExtComs(driver elevTypes.Drivers_ExtComs_s, orders elevTypes.Orders_ExtComs_s){    
    self.ExtComs.FloorChan          = driver.SensorChan
    self.ExtComs.StopButtonChan     = driver.StopButtonChan
    self.ExtComs.ObsChan            = driver.ObsChan    
    self.ExtComs.MotorChan          = driver.MotorChan
    self.ExtComs.DoorOpenChan       = driver.DoorOpenChan
    self.ExtComs.SetLightChan       = driver.SetLightChan
    self.ExtComs.SetFloorIndChan    = driver.SetFloorIndChan
    self.ExtComs.ElevPosRequest     = orders.ElevPosRequest
    self.ExtComs.NewOrdersChan      = orders.NewOrdersChan
    self.ExtComs.ExecdOrderChan     = orders.ExecdOrderChan
    self.ExtComs.ExecRequestChan    = orders.ExecRequestChan 
    self.ExtComs.ExecResponseChan   = orders.ExecResponseChan
    self.ExtComs.EmgTriggerdChan    = orders.EmgTriggerdChan
}


func (self *Fsm_s)start(){
    floor := <- self.ExtComs.FloorChan
    if floor == -1{
        self.ExtComs.MotorChan <- elevTypes.DOWN
    }
    for floor == -1{
        floor = <- self.ExtComs.FloorChan
    }   
    fmt.Println("               Found floor")
    //stop motor
    self.ExtComs.MotorChan <- elevTypes.NONE
    fmt.Println("                floor sendt")
    //update fsm vars
    self.state = IDLE
    self.lastFloor = floor
    self.lastDir = elevTypes.NONE
}


func (self *Fsm_s)update(){
    var event Event_t
    for{
        event =<- self.intComs.eventChan
        fmt.Println("               fmt.update: updating fsm[state][event]:", self.state, " ", event)
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
                    fmt.Println("               fsm.generate_events: reached new floor! lastFloor is now: ", floor) 
                    fsm.intComs.eventChan <- FLOOR_REACHED
               }
               
            case order:=<- fsm.ExtComs.NewOrdersChan:
                fmt.Println("               fsm: got new order on NewOrdersChan")
                
                go func() {fsm.intComs.eventChan <- NEW_ORDER}()
                fmt.Println("               fsm: sendt order on internal eventchan")
                
                go func() {fsm.intComs.newOrderChan <- order}()
                fmt.Println("               fsm: sendt order on internal newOrderChan")
                
            case execOrder:= <- fsm.ExtComs.ExecResponseChan:
                fmt.Println("               fsm.generate_events: got ExecResponse: ", execOrder)
                if execOrder{
                    fmt.Println("               fsm.generate_events: sending EXEC_ORDER on intComs.eventChan")
                    fsm.intComs.eventChan <- EXEC_ORDER
                }
                
            case <-fsm.intComs.timeoutChan:
                fmt.Println("               fsm.generate_eventes: TIMEOUT")
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
    fmt.Println("               fsm.start_down")
    self.ExtComs.MotorChan <- elevTypes.DOWN
    self.state = MOVING_DOWN
    self.lastDir = elevTypes.DOWN
    fmt.Println("               fsm: MOVING_DOWN\n")
}


func (self *Fsm_s)startUp(){
    fmt.Println("               fsm.start_up")
    self.ExtComs.MotorChan <- elevTypes.UP
    self.state = MOVING_UP
    self.lastDir = elevTypes.UP
    fmt.Println("               fsm: MOVING_UP\n")
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
