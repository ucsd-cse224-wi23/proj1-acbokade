package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"
	"math"
	"io"
	"bytes"
	"sort"

	"gopkg.in/yaml.v2"
)

const (
	PROT = "tcp"
)

type record struct {
	Key   [10]byte
	Value [90]byte
}

var serId int

var othersDataChannel []chan record = make([]chan record, 0)

type ServerConfigs struct {
	Servers []struct {
		ServerId int    `yaml:"serverId"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
	} `yaml:"servers"`
}

func readServerConfigs(configPath string) ServerConfigs {
	f, err := ioutil.ReadFile(configPath)

	if err != nil {
		log.Fatalf("could not read config file %s : %v", configPath, err)
	}

	scs := ServerConfigs{}
	err = yaml.Unmarshal(f, &scs)

	return scs
}

func checkError(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %s", err.Error())
	}
}

// Returns the address given the host name and port
func getIPAddress(host string, port string) string {
	return host + ":" + port
}

func connectToSocket(addr string) (net.Conn, error) {
	waitTime := time.Duration(250) // in ms
	// fmt.Println(serId, " Connecting to socket ", addr)
	for {
		conn, err := net.Dial(PROT, addr)
		// checkError(err)
		if err != nil {
			log.Fatalf("Connect to socket failed - Fatal error: %s %v", err.Error(), serId)
		}
		if err == nil {
			// fmt.Println(serId, " Connection to socket successful")
			return conn, nil
		}
		time.Sleep(waitTime)
	}
}

func acceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		// checkError(err)
		if err != nil {
			log.Fatalf("Accept connections failed - Fatal error: %s", err.Error())
		}
		if err == nil {
			// fmt.Println(serId, " Connection accepted ", conn)
			ch := make(chan record)
			othersDataChannel = append(othersDataChannel, ch)
			// Create a channel for the client
			go receiveData(conn, ch)
		}
	}
}

func receiveData(conn net.Conn, othersData chan<- record) {
	// fmt.Println(serId, " Receive data in progress ", conn)
	for {
		var key [10]byte
		var value [90]byte
		// // Read stream_complete boolean
		// checkError(err)
		
		var bytesToRead int = 0
		var buf[] byte
		var data[] byte
		data = make([]byte, 0)
		for {
			buf = make([]byte, 101 - bytesToRead)
			n, err := conn.Read(buf)
			// fmt.Println(serId, " value of n ", n, len(buf))
			// checkError(err)
			if err != nil {
				log.Fatalf("conn Read failed - Fatal error: %s %v", err.Error(), serId)
				continue
			}
			fmt.Println(serId, " Bytes read ", n)
			bytesToRead += n
			data = append(data, buf[:n]...)
			if bytesToRead >= 101 {
				// fmt.Println("bytes Read at the end ", bytesToRead)
				break
			}
		}
		// fmt.Println(serId, " buf content", buf)
		// fmt.Println(serId, " data content", data)
		stream_complete := (data[0] == 1)
		// fmt.Println(serId, " stream complete value ", stream_complete, data[0])
		if !stream_complete {
			fmt.Println("data length", len(data), len(buf), bytesToRead)
			copy(key[:], data[1:11])
			copy(value[:], data[11:101])
			rec := record{key, value}
			// fmt.Println(serId, " stream not complete ", key, value)
			othersData <- rec
		} else {
			fmt.Println(serId, " stream complete")
			// Close channel
			close(othersData)
			fmt.Println(serId, " Channel closed")
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
	serId = serverId
	if err != nil {
		log.Fatalf("Invalid serverId, must be an int %v", err)
	}

	// Read server configs from file
	scs := readServerConfigs(os.Args[4])

	// Read the port and address of own server
	myPort := scs.Servers[serverId].Port
	myAddr := scs.Servers[serverId].Host
	addr := getIPAddress(myAddr, myPort)
	/*
		Implement Distributed Sort
	*/
	/*
		Step 1: Create a mesh of TCP socket connections
	*/
	// Create a listener socket so that others can connect to
	listener, err := net.Listen(PROT, addr)
	// checkError(err)
	if err != nil {
		log.Fatalf("Net listen fail - Fatal error: %s", err.Error())
	}
	go acceptConnections(listener)
	time.Sleep(1000 * time.Millisecond)

	// Read input data and send to others
	inputFileName := os.Args[2]

	inputFile, err := os.Open(inputFileName)
	checkError(err)

	nServers := len(scs.Servers)
	nMSB := int(math.Log2(float64(nServers)))
	openConnections := make(map[int]net.Conn)
	// Establish connection with each other server
	for i := 0; i < nServers; i++ {
		if (i == serverId) {
			continue
		}
		peerPort := scs.Servers[i].Port
		peerAddr := scs.Servers[i].Host
		addr := getIPAddress(peerAddr, peerPort)
		conn, err := connectToSocket(addr)
		checkError(err)
		openConnections[i] = conn
	}

	// Declare records array which will store individual record
	records := []record{} 
	for {
		var key [10]byte
		var value [90]byte
		// Read first 10 bytes into key
		_, err := inputFile.Read(key[:])
		if err != nil {
			if err == io.EOF {
				fmt.Println(serId," Reached end of file while reading key")
				break
			}
			log.Println(err)
		}
		// Read the next 90 bytes into value
		_, err = inputFile.Read(value[:])
		if err != nil {
			if err == io.EOF {
				fmt.Println(serId, " Reached end of file while reading value")
				break
			}
			log.Println(err)
		}
		// keyToServerMapping, _ := strconv.Atoi(string(key[:nMSB]))
		firstByte := key[0]
		keyToServerMapping := int(firstByte) >> (8 - nMSB)
		// fmt.Println(serId, " keyToServerMapping ", keyToServerMapping)
		if keyToServerMapping == serverId {
			records = append(records, record{key, value})
			continue
		}
		conn, connExists := openConnections[keyToServerMapping]
		if !connExists && keyToServerMapping != serverId {
			// fmt.Println(serId, " First time connection")
			// Send the record to the server
			peerPort := scs.Servers[keyToServerMapping].Port
			peerAddr := scs.Servers[keyToServerMapping].Host
			addr := getIPAddress(peerAddr, peerPort)
			conn, err = connectToSocket(addr)
			checkError(err)
			openConnections[keyToServerMapping] = conn
		}
		if connExists {
			// fmt.Println(serId, " Connection already exists")
		}
		// Write 1 byte boolean
		var streamComplete byte = 0
		_, err = conn.Write([]byte{streamComplete})
		// checkError(err)
		if err != nil {
			log.Fatalf("Fail writing conn.write streamComplete Fatal error: %s", err.Error())
		}

		// Write 10 bytes key
		_, err = conn.Write([]byte(key[:]))
		// checkError(err)
		if err != nil {
			log.Fatalf("Fail writing conn.write key Fatal error: %s", err.Error())
		}

		// Write 90 bytes value
		_, err = conn.Write([]byte(value[:]))
		// checkError(err)
		// fmt.Println(serId, " Sent 101 bytes")
		if err != nil {
			log.Fatalf("Fail writing conn.write value Fatal error: %s", err.Error())
		}
		// fmt.Println(serId, "Sent 101 bytes")
	}
	// fmt.Println(serId, " Finished reading records in the file")
	time.Sleep(1000 * time.Millisecond)
	for _, conn := range openConnections {
		var key [10]byte
		var value [90]byte
		var streamComplete byte = 1
		_, err = conn.Write([]byte{streamComplete})

		// Write 10 bytes key
		_, err = conn.Write([]byte(key[:]))

		// Write 90 bytes value
		_, err = conn.Write([]byte(value[:]))
		conn.Close()
	}
	for i := 0; i < len(othersDataChannel); i++ {
		for rec := range othersDataChannel[i] {
			records = append(records, rec)
		}
		// fmt.Println(serId, " Completed reading channel ", i)
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
	fmt.Println(serId, "records sorted");
	// Write to output file

	// Read write file name
	writeFileName := os.Args[3]
	// Create output file
	outputFile, err := os.Create(writeFileName)
	if err != nil {
		log.Fatalf("Error creating output file - %v", err)
	}
	// Writing records to the output file
	for _, rec := range records {
		// Write key into the file
		_, err = outputFile.Write(rec.Key[:])
		if err != nil {
			log.Println(err)
		}
		// Write value into the file
		_, err = outputFile.Write(rec.Value[:])
		if err != nil {
			log.Println(err)
		}
	}
	// Closing the output file
	outputFile.Close()
}
