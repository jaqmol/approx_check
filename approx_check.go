package main

import (
	"bufio"
	"encoding/json"
	"time"

	"github.com/jaqmol/approx/errormsg"
	"github.com/jaqmol/approx/processorconf"
)

// NewApproxCheck ...
func NewApproxCheck(conf *processorconf.ProcessorConf) *ApproxCheck {
	errMsg := &errormsg.ErrorMsg{Processor: "approx_test"}
	modeEnv := conf.Envs["MODE"]
	var mode Mode
	if "produce" == modeEnv {
		mode = ModeProduce
	} else if "consume" == modeEnv {
		mode = ModeConsume
	} else {
		errMsg.LogFatal(nil, "Test expects env MODE to be either produce or consume, but got %v", modeEnv)
	}
	return &ApproxCheck{
		errMsg:    errMsg,
		conf:      conf,
		output:    conf.Outputs[0],
		input:     conf.Inputs[0],
		mode:      mode,
		idCounter: 0,
		date:      time.Now(),
	}
}

// ApproxCheck ...
type ApproxCheck struct {
	errMsg    *errormsg.ErrorMsg
	conf      *processorconf.ProcessorConf
	output    *bufio.Writer
	input     *bufio.Reader
	mode      Mode
	idCounter int
	date      time.Time
}

// Mode ...
type Mode int

// Mode Types
const (
	ModeProduce Mode = iota
	ModeConsume
)

// Start ...
func (a *ApproxCheck) Start() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {
		a.idCounter++

		msg := a.nextDayReq()
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			a.errMsg.Log(&a.idCounter, "Error marshalling request message: %v", err.Error())
			return
		}

		msgBytes = append(msgBytes, '\n')
		_, err = a.output.Write(msgBytes)
		if err != nil {
			a.errMsg.Log(&a.idCounter, "Error writing request message to output: %v", err.Error())
			return
		}

		err = a.output.Flush()
		if err != nil {
			a.errMsg.Log(&a.idCounter, "Error flushing written message to output: %v", err.Error())
			return
		}
	}
}

func (a *ApproxCheck) nextDayReq() *TimeReq {
	return &TimeReq{
		JSONRPC: "2.0",
		ID:      a.idCounter,
		Method:  "NextDay",
		Params: Params{
			Day:     a.date.Day(),
			Month:   int(a.date.Month()),
			Year:    a.date.Year(),
			Weekday: a.date.Weekday().String(),
		},
	}
}

type TimeReq struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  Params `json:"params"`
}

// Params ...
type Params struct {
	Day     int    `json:"day"`
	Month   int    `json:"month"`
	Year    int    `json:"year"`
	Weekday string `json:"weekday"`
}
