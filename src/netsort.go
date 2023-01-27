package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

// Constants
const (
	PROT = "tcp"
)

// Struct for record
type record struct {
	Key   [10]byte
	Value [90]byte
}

var nServers int // number of servers as defined in the config.yaml file
var openClientConnections []net.Conn  // open connections with the clients

// Contains the channel corresponding to each other node where
// this server intend to receive the data from
var othersDataChannel []chan record = make([]chan record, 0) 

type ServerConfigs struct {
	Servers []struct {
		ServerId int    `yaml:"serverId"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
	} `yaml:"servers"`
}

// Read the config yaml file and parse it
func readServerConfigs(configPath string) ServerConfigs {
	f, err := ioutil.ReadFile(configPath)

	if err != nil {
		log.Fatalf("could not read config file %s : %v", configPath, err)
	}

	scs := ServerConfigs{}
	err = yaml.Unmarshal(f, &scs)

	return scs
}

func checkErrorWithExit(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %s", err)
	}
}

func checkErrorWithoutExit(err error) {
	if err != nil {
		log.Printf("Error: %s\n", err)
	}
}

// Returns the address given the host name and port
func getIPAddress(host string, port string) string {
	return host + ":" + port
}

// Connects to the socket using address provided and
// waits for 250 ms before retrying
func connectToSocket(addr string) (net.Conn, error) {
	waitTime := time.Duration(250) // in ms
	for {
		conn, err := net.Dial(PROT, addr)
		checkErrorWithoutExit(err)
		if err == nil {
			return conn, nil
		}
		time.Sleep(waitTime)
	}
}

// Makes the listener to accept to other nodes and
// appends a channel for that node to othersDataChannel
func acceptConnections(listener net.Listener) {
	for {
		if len(openClientConnections) == nServers - 1 {
			break
		}
		conn, err := listener.Accept()
		checkErrorWithExit(err)
		openClientConnections = append(openClientConnections, conn)
		if err == nil {
			ch := make(chan record)
			othersDataChannel = append(othersDataChannel, ch)
			// Create a channel for the client
			go receiveData(conn, ch)
		}
	}
}

// Receives data from the connection provided and append
// it to the given channel
func receiveData(conn net.Conn, othersData chan<- record) {
	for {
		var key [10]byte
		var value [90]byte
		var bytesToRead int = 0
		var buf []byte
		var data []byte
		data = make([]byte, 0)
		for {
			// Make a buf of size 101-bytesToRead to ensure exactly 101 bytes
			// are read and then used to create a record object
			buf = make([]byte, 101-bytesToRead)
			n, err := conn.Read(buf)
			checkErrorWithExit(err) // exit since data if not read will output incorrect result
			bytesToRead += n
			data = append(data, buf[:n]...)
			if bytesToRead >= 101 {
				break
			}
		}
		streamComplete := (data[0] == 1)
		if !streamComplete {
			copy(key[:], data[1:11])
			copy(value[:], data[11:101])
			rec := record{key, value}
			othersData <- rec
		} else {
			// Close channel
			close(othersData)
			break
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) != 5 {
		log.Fatal("Usage : ./netsort {serverId} {inputFilePath} {outputFilePath} {configFilePath}")
	}

	// What is my serverId
	serverId, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid serverId, must be an int %v", err)
	}

	// Read server configs from file
	scs := readServerConfigs(os.Args[4])
	sort.Slice(scs.Servers, func(i, j int) bool {
		return scs.Servers[i].ServerId <= scs.Servers[j].ServerId
	})
	nServers = len(scs.Servers)

	// Read the port and address of own server
	myPort := scs.Servers[serverId].Port
	myAddr := scs.Servers[serverId].Host
	addr := getIPAddress(myAddr, myPort)
	/*
		Implement Distributed Sort
	*/
	// Create a listener socket so that others can connect to
	listener, err := net.Listen(PROT, addr)
	checkErrorWithExit(err)
	defer listener.Close()
	go acceptConnections(listener)
	// Wait for some time so that other nodes can establish the listener socket
	// and accepting connections
	time.Sleep(1000 * time.Millisecond)

	// Read input data and send to others
	inputFileName := os.Args[2]

	inputFile, err := os.Open(inputFileName)
	checkErrorWithExit(err)

	nMSB := int(math.Log2(float64(nServers)))
	openConnections := make(map[int]net.Conn)
	// Establish connection with each other server
	for i := 0; i < nServers; i++ {
		if i == serverId {
			continue
		}
		peerPort := scs.Servers[i].Port
		peerAddr := scs.Servers[i].Host
		addr := getIPAddress(peerAddr, peerPort)
		conn, _ := connectToSocket(addr)
		openConnections[i] = conn
		defer conn.Close()
	}
	// Wait for some time so that other nodes can establish the listener socket
	// and accepting connections
	time.Sleep(1000 * time.Millisecond)

	// Declare records array which will store individual record
	records := []record{}
	for {
		var key [10]byte
		var value [90]byte
		// Read first 10 bytes into key
		_, err := inputFile.Read(key[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			checkErrorWithExit(err)
		}
		// Read the next 90 bytes into value
		_, err = inputFile.Read(value[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			checkErrorWithExit(err)
		}
		firstByte := key[0]
		keyToServerMapping := int(firstByte) >> (8 - nMSB)
		if keyToServerMapping == serverId {
			records = append(records, record{key, value})
			continue
		}
		conn, connExists := openConnections[keyToServerMapping]
		if !connExists && keyToServerMapping != serverId {
			peerPort := scs.Servers[keyToServerMapping].Port
			peerAddr := scs.Servers[keyToServerMapping].Host
			addr := getIPAddress(peerAddr, peerPort)
			conn, _ = connectToSocket(addr)
			openConnections[keyToServerMapping] = conn
		}
		// Write 1 byte boolean
		var streamComplete byte = 0
		_, err = conn.Write([]byte{streamComplete})
		checkErrorWithExit(err)

		// Write 10 bytes key
		_, err = conn.Write([]byte(key[:]))
		checkErrorWithExit(err)

		// Write 90 bytes value
		_, err = conn.Write([]byte(value[:]))
		checkErrorWithExit(err)
	}

	time.Sleep(1000 * time.Millisecond)

	for _, conn := range openConnections {
		var key [10]byte
		var value [90]byte
		var streamComplete byte = 1
		// Write 1 byte to indicate stream completion
		_, err = conn.Write([]byte{streamComplete})
		checkErrorWithExit(err)

		// Write 10 bytes key
		_, err = conn.Write([]byte(key[:]))
		checkErrorWithExit(err)

		// Write 90 bytes value
		_, err = conn.Write([]byte(value[:]))
		checkErrorWithExit(err)
		// defer conn.Close()
	}

	// Append records from all the sender nodes' channels
	for i := 0; i < len(othersDataChannel); i++ {
		for rec := range othersDataChannel[i] {
			records = append(records, rec)
		}
	}

	// Close input file
	inputFile.Close()

	// Sort the data
	// Custom comparator for sorting records array by key
	sort.Slice(records, func(i, j int) bool {
		// Sort the two records by the key
		isLessThan := bytes.Compare(records[i].Key[:], records[j].Key[:])
		if isLessThan <= 0 {
			return true
		}
		return false
	})

	// Write to output file
	// Read write file name
	writeFileName := os.Args[3]
	// Create output file
	outputFile, err := os.Create(writeFileName)
	checkErrorWithExit(err)
	// Writing records to the output file
	for _, rec := range records {
		// Write key into the file
		_, err = outputFile.Write(rec.Key[:])
		checkErrorWithExit(err)
		// Write value into the file
		_, err = outputFile.Write(rec.Value[:])
		checkErrorWithExit(err)
	}
	// Closing the output file
	outputFile.Close()
}
