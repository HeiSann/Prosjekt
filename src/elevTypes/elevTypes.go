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

type Net_ExtComs_s struct{
   Dummy int
}

type Orders_ExtComs_s struct{
   Dummy int
}

type Drivers_ExtComs_s struct{
	/* Channels initialized in driver */
   ButtonChan 			<-chan Button
	SensorChan 			<-chan int
	StopButtonChan 	<-chan bool
	ObsChan 				<-chan bool
	MotorChan 			chan<- Direction_t
	SetLightChan 		chan<- int
	SetFloorIndChan 	chan<- int
	DoorOpenChan      chan<- bool
}


type Fsm_ExtComs_s struct{
	/* Channels initialized in fsm */   
   OrderExdChan     	chan Order_t  //fsm -> orders
	/* Channels from driver */
   ButtonChan        <-chan Button
   FloorChan         <-chan int
   StopButtonChan    <-chan bool
   ObsChan           <-chan bool
   MotorChan         chan<- Direction_t
   DoorOpenChan      chan<- bool
   SetLightChan      chan<- Light_t
   SetFloorIndChan   chan<- int 
	/* Channels from orders*/
   NewOrdersChan    <-chan Order_t  //orders -> fsm
   EmgTriggerChan   <-chan bool     //orders -> fsm
}


type Message struct{
	To string
	From string //ipAdr
	Msg_type string //order, deadElev, auction, connect to me
	Payload string
}

   





