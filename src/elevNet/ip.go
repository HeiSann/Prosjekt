package elevNet

import (
	"net"
	"strings"
)
func GetMyIP() string{
	allIPs, err := net.InterfaceAddrs()
	if err != nil {
	    return ""
	}
			
	uncutIP := allIPs[1].String()
	IP := strings.Split(uncutIP, "/")		
	myIP := IP[0]

	return myIP
}

func GetBroadcastIP(myIP string) string{
	temp :=strings.Split(myIP,".")
	broadCastIP :=temp[0]+"."+temp[1]+"."+temp[2]+"."+"255"
	return broadCastIP
	

}



