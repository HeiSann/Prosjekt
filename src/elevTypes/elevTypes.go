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
	/* Channels from comsManager */
	OrderToNetChan  	chan<- Order_t	
	NetToOrderNew		<-chan Order_t
	RequestScore		chan<- Order_t
	RespondScore		<-chan Order_t
	/* Channels initialized in driver */
   NewOrdersChan    	<-chan Order_t 
   OrderUpdatedChan	<-chan Order_t 
	OrderExecdChan  	chan<- Order_t	
	StopRequestChan  	chan<- Order_t		
   EmgTriggerdChan  	<-chan bool
}

type Drivers_ExtComs_s struct{
	/* Channels initialized in driver */
   ButtonChan 			<-chan Button
	SensorChan 			<-chan int
	StopButtonChan 	<-chan bool
	ObsChan 				<-chan bool
	MotorChan 			chan<- Direction_t
	SetLightChan 		chan<- Light_t
	SetFloorIndChan 	chan<- int
	DoorOpenChan      chan<- bool
}


type Fsm_ExtComs_s struct{
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
	NewOrdersChan    	<-chan Order_t		//order sends only when requested, or new order in empty queue
	OrderExecChan		<-chan Order_t 
	OrderExecdChan  	chan<- Order_t	
	StopRequestChan  	chan<- Order_t		
   EmgTriggerdChan  	<-chan bool     	
}


type Message struct{
	To string
	From string //ipAdr
	Msg_type string //order, deadElev, auction, connect to me
	Payload string
}

   





