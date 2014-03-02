package comsManager
import( "elevTypes"
		  )
  



//func InitElevList()[]Elev{

//    ElevList:=make([]Elev,999)
//}

// Mail ElevPackage Delivery Message 



type ExternalChan_s struct{
	send chan elevTypes.Message
}
    
type InternalChan_s struct{

}

type ComsManager_s struct{
	ExtComs ExternalChan_s
	intComs InternalChan_s
	
}
