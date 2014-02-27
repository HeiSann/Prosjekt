package message
import "time"
import "strconv"


import "strings"

const SIZE = 1024

type Message struct{
	To string
	From string //ipAdr
	Msg_type string //order, deadElev, auction, connect to me
	Payload string
}




func ConstructMessage(to string, from string, typ string, msg string) Message{
	return Message{to,from,typ,msg}
}

//not generic, could use reflect...
func Message2bytestream (m Message) []byte{
	msg := m.To +"~"+m.From +"~"+ m.Msg_type +"~"+ m.Payload
	return []byte(msg+"\x00")
}

//not generic, could use reflect..
func Bytestream2message(m []byte) Message{
	msg_string := string(m[:])
	msg_array := strings.Split(msg_string, "~")
	msg := Message{msg_array[0], msg_array[1], msg_array[2], msg_array[3]}
	return msg
}

func ConstructPing(ipTo string, ipFrom string)Message{
    timestamp:=time.Now().Unix()
    stimestamp:=strconv.FormatInt(timestamp,10)
    return Message{ipTo, ipFrom, "PING", stimestamp}
}

func Construct(ipTo string, ipFrom string)Message{
    
return Message{ipTo, ipFrom, "SDfs","msg"}
}
