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


type Drivers_s struct{
   ButtonChan chan Button
	SensorChan chan int
	MotorChan chan Direction_t
	StopButtonChan chan bool
	ObsChan chan bool
}


type Net_s struct{
   Dummy int
}

type Order_t struct{
   floor       int
   direction   Direction_t   
   status      bool
}

type Light_t struct{
   floor       int
   direction   Direction_t   
   status      bool
}


   
   





