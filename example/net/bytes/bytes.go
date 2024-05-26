package main

import (
	"fmt"
	"github.com/joe-at-startupmedia/xipc"
	"github.com/joe-at-startupmedia/xipc/net"
	"log"
)

const maxRequestTickNum = 10
const queue_name = "goqr_example_bytes"

var mqr xipc.IMqResponder
var mqs xipc.IMqRequester
var config = net.QueueConfig{
	Name:             queue_name,
	ClientTimeout:    0,
	ClientRetryTimer: 0,
}

func main() {
	resp_c := make(chan int)
	go responder(resp_c)
	//wait for the responder to create the posix_mq files
	xipc.Sleep()
	request_c := make(chan int)
	go requester(request_c)
	<-resp_c
	<-request_c

	mqs.CloseRequester()
	mqr.CloseResponder()
	//gives time for deferred functions to complete
	xipc.Sleep()
}

func responder(c chan int) {

	mqr = net.NewResponder(&config)
	xipc.Sleep()

	defer func() {
		log.Println("Responder: finished")
		c <- 0
	}()
	if mqr.HasErrors() {
		log.Printf("Responder: could not initialize: %s", mqr.Error())
		c <- 1
		return
	}

	count := 0
	for {
		count++

		if err := mqr.HandleRequest(handleMessage); err != nil {
			log.Printf("Responder: error handling request: %s\n", err)
		} else {
			log.Printf("Processed response")
		}

		if count >= maxRequestTickNum {
			break
		}
	}
}

func requester(c chan int) {

	mqs = net.NewRequester(&config)

	defer func() {
		log.Println("Requester: finished and closed")
		c <- 0
	}()
	if mqs.HasErrors() {
		log.Printf("Requester: could not initialize: %s", mqs.Error())
		c <- 1
		return
	}
	xipc.Sleep()

	count := 0
	for {

		count++

		request := fmt.Sprintf("Hello, World : %d\n", count)
		if err := mqs.Request([]byte(request)); err != nil {
			log.Printf("Requester: error requesting request: %s\n", err)
			continue
		}

		log.Printf("Requester: sent a new request: %s", request)

		msg, err := mqs.WaitForResponse()

		if err != nil {
			log.Printf("Requester: error getting response: %s\n", err)
			continue
		}

		log.Printf("Requester: got a response: %s\n", msg)

		if count >= maxRequestTickNum {
			break
		}
	}
}

func handleMessage(request []byte) (processed []byte, err error) {
	return []byte(fmt.Sprintf("I recieved request: %s\n", request)), nil
}
