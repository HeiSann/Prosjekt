package elevTypes

const N_FLOORS = 4

type Direction_t int 

const (
    UP Direction_t = iota
    DOWN
    NONE
    N_DIR
)

type Button struct{
	Floor int    
	Dir Direction_t          
}

type Light_t struct{
   Floor       int
   Direction   Direction_t   
   Set      	bool
}

type Order_t struct{
   Floor       int
   Direction   Direction_t   
   Status      bool
}

type Drivers_ExtComs_s struct{
   ButtonChan chan Button
	SensorChan chan int
	MotorChan chan Direction_t
	StopButtonChan chan bool
	ObsChan chan bool
}

type Net_ExtComs_s struct{
   Dummy int
}

type Orders_ExtComs_s struct{
   Dummy int
}

type Fsm_ExtComs_s struct{
   ButtonChan        chan Button
   FloorChan         chan int
   StopButtonChan    chan bool
   ObsChan           chan bool
   
   MotorChan         chan Direction_t
   DoorOpenChan      chan bool
   SetLightChan      chan Light_t
   FloorIndChan      chan int 
   
   OrderExdChan     chan Order_t  //fsm -> orders
   NewOrdersChan    chan Order_t  //orders -> orders
   EmgTriggerChan   chan bool     //orders -> fsm
}


   
   





