package elevWatchdog

import(
   "fmt"
)

/*

System design allows for reinitializing of each module seperatly if any unexpected deadlocks should occur. 
In order to loose as little state information as possible each module should have been implimented with a makeDouble-function 
so that important information such as orders or last know floor or direction is preserved

All important selects should report to the watchdog in a timely fashion. Whenever a module fails to report, 
the watchdog calls the makeDouble-function of that module (already implemented in both orders and fsm), and replaces the original (failing) 
module member struct in the the original Elevator struct with the newly initiated one. 

If however the reinited module quickly fails again (probably due to its copied information), the entire system is to be re inited from scratch


*/