package elevTypes

const N_FLOORS = 4
const N_DIR = 3
const DOOR_OPEN_TIME = 3 //millisec
const SLOW_DOWM_MUTHA_FUKKA = 20


type Direction_t int 
const (
    UP Direction_t = iota
    DOWN
    NONE
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
   Active      bool
}

type ElevPos_t struct{
   Floor       int
   Direction   Direction_t   
   Status      bool
}

type Net_ExtComs_s struct{
   RecvMsg chan Message
	SendMsg chan Message  
	SendBcast chan Message
	HeartbeatMsg chan Message
	SendMsgToAll chan Message
	DeadElev chan string
	NewElev chan string
	FailedTcpMsg chan Message
}

type ComsManager_ExtComs_s struct{
	/* inited in self */
	send chan Message
	//chan to order init here
	AuctionOrder chan Order_t //external oder in elevator. This will star auction
	RequestCost chan Order_t
	RecvCost chan int
	AddOrder chan Order_t
	SendOrderUpdate chan Order_t
	RecvOrderUpdate chan Message
	AuctionDeadElev chan string
	CheckNewElev chan Message
	UpdateElevInside chan Message
      
	/*inited in net*/
	RecvMsg chan Message
	SendMsg chan Message  
	SendBcast chan Message
	HeartbeatMsg chan Message
	SendMsgToAll chan Message
	DeadElev chan string
	NewElev chan string
	FailedTcpMsg chan Message
}

type Orders_ExtComs_s struct{
	/* Channels initialized in orders */
	ElevPosRequest      chan ElevPos_t
	NewOrdersChan    	chan Order_t 
	ExecdOrderChan  	chan ElevPos_t	
	ExecRequestChan  	chan ElevPos_t	
	ExecResponseChan	chan bool	
	EmgTriggerdChan  	chan bool
	/* Channels from comsManager */
	AuctionOrder		chan Order_t
	RequestScoreChan	chan Order_t
	RespondScoreChan	chan int
	AddOrder 			chan Order_t	
	SendOrderUpdate 	chan Order_t
	RecvOrderUpdate 	chan Message
	AuctionDeadElev   	chan string
	CheckNewElev		chan Message
	UpdateElevInside 	chan Message
	/* Channels from driver */
	ButtonChan        <-chan Button
	SetLightChan      chan<- Light_t
}

type Drivers_ExtComs_s struct{
	/* Channels initialized in driver */
	ButtonChan 		<-chan Button
	SensorChan 		<-chan int
	StopButtonChan 	<-chan bool
	ObsChan 		<-chan bool
	MotorChan 		chan<- Direction_t
	SetLightChan 	chan<- Light_t
	SetFloorIndChan chan<- int
	DoorOpenChan    chan<- bool
}

type Fsm_ExtComs_s struct{
	/* Channels from driver */
	FloorChan         <-chan int
	StopButtonChan    <-chan bool
	ObsChan           <-chan bool
	MotorChan         chan<- Direction_t
	DoorOpenChan      chan<- bool
	SetLightChan      chan<- Light_t
	SetFloorIndChan   chan<- int 
	/* Channels from orders*/
	ElevPosRequest      chan ElevPos_t
	NewOrdersChan		chan Order_t 
	ExecdOrderChan		chan ElevPos_t	
	ExecRequestChan		chan ElevPos_t	
	ExecResponseChan	chan bool	
	EmgTriggerdChan		chan bool   	
}


type Message struct{
	To string
	From string //ipAdr
	Type string //order, deadElev, auction, connect to me
	Payload string
	Cost int
	Order Order_t
}

   





