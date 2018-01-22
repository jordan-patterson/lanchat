package main 

import(
	"fmt"
	"net"
	"time"
	"log"
	"os"
	"bufio"
	"strings"
	"flag"
)
type Manager struct{
	Clients map[Client]bool
	Register chan Client
	Unregister chan Client
}
type Client struct{
	Id int
	Connection net.Conn
}
func (manager *Manager) Start(){
	for{
		select{
			case client:=<-manager.Register:
				manager.Clients[client]=true
				//following to be changed to broadcast
				manager.Broadcast("New Client has joined the chatroom",-1)
				go manager.HandleClient(client)
			case client:=<-manager.Unregister:
				//delete(manager.Clients,client)
				manager.Clients[client]=false
				//following to be changed to broadcast
				manager.Broadcast("A client connection has been terminated",-1)
		}
	}
}
func StartServer(){
	ctime:=time.Now()
	ip:=GetLocalIP()
	socket,err:=net.Listen("tcp",ip+":8000")
	fmt.Println("\tStarted server at "+ctime.String()+" on "+ip+":8000")
	if err!=nil{
		log.Fatal(err)
	}
	chatroom:=Manager{
		Clients:make(map[Client]bool),
		Register:make(chan Client),
		Unregister:make(chan Client),
	}
	go chatroom.Start()

	for {
		//waits for incoming connection
		conn,_:=socket.Accept()
		/*
		creates new client struct from connection 
		and sends it through register channel for handling
		*/
		cli:=Client{
			Id:len(chatroom.Clients)+1,
			Connection:conn,
		}
		chatroom.Register<-cli
	}
}
func (manager *Manager) HandleClient(client Client){
	//recieves messages from client after which are broadcasted
	for{
		msg:=make([]byte,2048)
		length,err:=client.Connection.Read(msg)
		if err!=nil{
			fmt.Println(err.Error())
			manager.Unregister<-client
			return
		}
		if length>0{
			go manager.Broadcast(string(msg),client.Id)
		}
	}
}
func (manager *Manager) Broadcast(msg string,ignore int){
	bytes:=[]byte(msg)
	ctime:=time.Now()
	fmt.Println(msg+" "+ctime.String())
	for client,ok := range manager.Clients{
		if client.Id!=ignore && ok{
			_,err:=client.Connection.Write(bytes)
			if err!=nil{
				manager.Unregister<-client
				client.Connection.Close()
			}
		}
	} 
}
func StartClient(){
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter the IP address on the network in which the server is running: ")
	scanner.Scan()
	ip:=scanner.Text()
	conn,err:=net.Dial("tcp",ip+":8000")
	if err!=nil{
		log.Fatal(err)
	}
	fmt.Print("Successfully connected to the chatrooom. Type 'q' to quit anytime.\nEnter your name: ")
	scanner.Scan()
	go recieve(conn)
	alias:=scanner.Text()
	for scanner.Scan(){
		msg:=scanner.Text()
		if strings.ToLower(msg)=="q"{
			conn.Close()
			os.Exit(3)
		}else{
			msg:=alias+" : "+msg
			_,err:=conn.Write([]byte(msg))
			if err!=nil{
				log.Fatal(err)
			}
		}
	}
}
func recieve(conn net.Conn){
	for{
		bytes:=make([]byte,2048)
		length,err:=conn.Read(bytes)
		if err!=nil{
			log.Fatal(err)
		}
		if length>0{
			fmt.Println(string(bytes))
		}
	}
}
func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}
func main(){
	mode := flag.String("mode","server","run program as server or client")
	flag.Parse()
	if strings.ToLower(*mode)=="server"{
		StartServer()
	}else{
		StartClient()
	}
}
