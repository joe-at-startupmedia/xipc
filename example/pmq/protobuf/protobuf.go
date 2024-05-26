package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/joe-at-startupmedia/posix_mq"
	"github.com/joe-at-startupmedia/xipc/example/protos"
	"github.com/joe-at-startupmedia/xipc/pmq"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

const maxRequestTickNum = 10

const queue_name = "pmqr_example_protobuf"

var owner = pmq.Ownership{
	//Username: "nobody", //uncomment to test ownership handling and errors
}

func main() {
	resp_c := make(chan int)
	go responder(resp_c)
	//wait for the responder to create the posix_mq files
	time.Sleep(1 * time.Second)
	request_c := make(chan int)
	go requester(request_c)
	<-resp_c
	<-request_c
	//gives time for deferred functions to complete
	time.Sleep(2 * time.Second)
}

func responder(c chan int) {
	config := pmq.QueueConfig{
		Name:  queue_name,
		Flags: posix_mq.O_RDWR | posix_mq.O_CREAT,
	}
	mqr := pmq.NewResponder(&config, &owner)
	defer func() {
		pmq.UnlinkResponder(mqr)
		fmt.Println("Responder: finished and unlinked")
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
		if err := handleCmdRequest(mqr); err != nil {
			fmt.Printf("Responder: error handling request: %s\n", err)
			continue
		}

		fmt.Println("Responder: Sent a response")

		if count >= maxRequestTickNum {
			break
		}
	}
}

func requester(c chan int) {
	mqs := pmq.NewRequester(&pmq.QueueConfig{
		Name: queue_name,
	}, &owner)
	defer func() {
		pmq.CloseRequester(mqs)
		fmt.Println("Requester: finished and closed")
		c <- 0
	}()
	if mqs.HasErrors() {
		log.Printf("Requester: could not initialize: %s", mqs.Error())
		c <- 1
		return
	}

	count := 0
	for {
		count++
		cmd := &protos.Cmd{
			Name: "restart",
			Arg1: fmt.Sprintf("%d", count), //using count as the id of the process
			ExecFlags: &protos.ExecFlags{
				User: "nonroot",
			},
		}
		if err := requestUsingCmd(mqs, cmd); err != nil {
			fmt.Printf("Requester: error requesting request: %s\n", err)
			continue
		}

		fmt.Printf("Requester: sent a new request: %s \n", cmd.String())

		cmdResp, err := waitForCmdResponse(mqs)

		if err != nil {
			fmt.Printf("Requester: error getting response: %s\n", err)
			continue
		}

		fmt.Printf("Requester: got a response: %s\n", cmdResp.ValueStr)
		//fmt.Printf("Requester: got a response: %-v\n", msg)

		if count >= maxRequestTickNum {
			break
		}
	}
}

func requestUsingCmd(mqs *pmq.MqRequester, req *protos.Cmd) error {
	if len(req.Id) == 0 {
		req.Id = uuid.NewString()
	}
	pbm := proto.Message(req)
	return mqs.RequestUsingProto(&pbm)
}

func waitForCmdResponse(mqs *pmq.MqRequester) (*protos.CmdResp, error) {
	mqResp := &protos.CmdResp{}
	_, err := mqs.WaitForProto(mqResp)
	if err != nil {
		return nil, err
	}
	return mqResp, err
}

// handleCmdRequest provides a concrete implementation of HandleRequestFromProto using the local Cmd protobuf type
func handleCmdRequest(mqr *pmq.MqResponder) error {

	cmd := &protos.Cmd{}

	return mqr.HandleRequestFromProto(cmd, func() (processed []byte, err error) {

		cmdResp := protos.CmdResp{}
		cmdResp.Id = cmd.Id
		cmdResp.ValueStr = fmt.Sprintf("I recieved request: %s(%s) - %s\n", cmd.Name, cmd.Id, cmd.Arg1)

		data, err := proto.Marshal(&cmdResp)
		if err != nil {
			return nil, fmt.Errorf("marshaling error: %w", err)
		}

		return data, nil
	})
}
