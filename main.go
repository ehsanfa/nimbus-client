package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
)

type input struct {
	action action
	args   []string
}

type action int64

const (
	UNDEFINED action = iota
	GET
	SET
)

type GetRequest struct {
	Key string
}

type GetResponse struct {
	Value string
	Ok    bool
	Error string
}

type SetRequest struct {
	Key   string
	Value string
}

type SetResponse struct {
	Ok    bool
	Error string
}

func readInput(i chan input, exit chan struct{}) {
	reader := bufio.NewReader(os.Stdin)
	txt, _ := reader.ReadString('\n')
	txt = strings.TrimSuffix(strings.ToLower(txt), "\n")
	slices := strings.Split(txt, " ")
	if len(slices) == 0 {
		i <- input{UNDEFINED, []string{}}
		return
	}
	var action action
	a := strings.ToLower(slices[0])
	switch a {
	case "get":
		action = GET
	case "set":
		action = SET
	case "exit":
		exit <- struct{}{}
	default:
		action = UNDEFINED
	}
	i <- input{action, slices[1:]}
}

func getRequest(client *rpc.Client, args []string) (string, error) {
	var resp GetResponse
	err := client.Call("Cacher.Get", GetRequest{args[0]}, &resp)
	if err != nil {
		return "", err
	}
	if resp.Ok != true {
		return "", errors.New(resp.Error)
	}
	return resp.Value, nil
}

func setRequest(client *rpc.Client, args []string) error {
	if len(args) < 2 {
		return errors.New("Too few arguments")
	}
	var resp SetResponse
	err := client.Call("Cacher.Set", SetRequest{args[0], args[1]}, &resp)
	if err != nil {
		return err
	}
	if resp.Ok != true {
		return errors.New(resp.Error)
	}
	return nil
}

func handleInput(client *rpc.Client, req input) {
	switch req.action {
	case GET:
		resp, err := getRequest(client, req.args)
		fmt.Println(resp, err)
	case SET:
		err := setRequest(client, req.args)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("OK")
		}
	case UNDEFINED:
		fmt.Println("Undefined action")
	}
}

func main() {
	client, err := rpc.DialHTTP("tcp", "localhost:9022")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	input := make(chan input)
	exit := make(chan struct{})

	fmt.Println("Enter Command")
	go func() {
		for {
			readInput(input, exit)
		}
	}()

client:
	for {
		select {
		case <-exit:
			break client
		case i := <-input:
			handleInput(client, i)
		}

	}

}
