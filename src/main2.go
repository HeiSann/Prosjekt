package main

import(
   "fmt"
   "elevTypes"
   "messages"
   "elevNet"
   "comsManager"
   "elevDrivers"
   "elevOrders"
   "elevFSM"
)

type Elevator struct{
	driver      Drivers_s
	net         Net_s
	coms        ComsManager_s
	orders      Orders_s
	fsm         Fsm_s
}

func main(){
   var net = elevNet.init()
   var coms = comsManager(net.ExtChan)
   var drivers = elevDrivers.init()
   var orders = elevOrders.init(drivers.ExtChan, coms.ExtChan)
   var fsm = elevFSM.init(drivers.ExtComs, orders.ExtComs)
   
   var Elev = elevTypes.Elevator{net, coms, driver, orders, fsm}
       
   for{}
    
}
