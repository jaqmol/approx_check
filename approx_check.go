package main

import (
	"bufio"
	"io"
	"time"

	"github.com/jaqmol/approx/axmsg"
	"github.com/jaqmol/approx/processorconf"
)

// NewApproxCheck ...
func NewApproxCheck(conf *processorconf.ProcessorConf) *ApproxCheck {
	errMsg := &axmsg.Errors{Source: "approx_check"}

	modeEnv := conf.Envs["MODE"]
	var mode Mode
	switch modeEnv {
	case "produce":
		mode = ModeProduce
	case "consume":
		mode = ModeConsume
	default:
		errMsg.LogFatal(nil, "Test expects env MODE to be either produce or consume, but got %v", modeEnv)
	}

	var speed Speed
	if ModeProduce == mode {
		speedEnv, ok := conf.OptionalEnv("SPEED")
		if !ok {
			errMsg.LogFatal(nil, "Test expects env SPEED, if MODE is produce")
		}
		switch speedEnv {
		case "untethered":
			speed = SpeedUntethered
		case "fast":
			speed = SpeedFast
		case "moderate":
			speed = SpeedModerate
		case "slow":
			speed = SpeedSlow
		default:
			errMsg.LogFatal(nil, "Test expects env SPEED to be either untethered, fast, moderate or slow, but got %v", modeEnv)
		}
	}

	return &ApproxCheck{
		errMsg:    errMsg,
		conf:      conf,
		output:    axmsg.NewWriter(conf.Outputs[0]),
		input:     conf.Inputs[0],
		mode:      mode,
		speed:     speed,
		idCounter: 0,
		date:      time.Now(),
	}
}

// ApproxCheck ...
type ApproxCheck struct {
	errMsg    *axmsg.Errors
	conf      *processorconf.ProcessorConf
	output    *axmsg.Writer
	input     *bufio.Reader
	mode      Mode
	speed     Speed
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

// Speed ...
type Speed int

// Speed Types
const (
	SpeedUntethered Speed = iota
	SpeedFast
	SpeedModerate
	SpeedSlow
)

// Start ...
func (a *ApproxCheck) Start() {
	if ModeProduce == a.mode {
		if SpeedUntethered == a.speed {
			a.startUntetheredProduce()
		} else {
			a.startTetheredProduce()
		}
	} else if ModeConsume == a.mode {
		a.startConsume()
	}
}

func (a *ApproxCheck) startUntetheredProduce() {
	for {
		a.produceNext()
	}
}

func (a *ApproxCheck) startTetheredProduce() {
	ticker := time.NewTicker(a.duration())
	for range ticker.C {
		a.produceNext()
	}
}

func (a *ApproxCheck) produceNext() {
	msg := a.nextDateAction()
	err := a.output.Write(msg)
	if err != nil {
		a.errMsg.Log(&a.idCounter, "Error writing request message to output: %v", err.Error())
		return
	}
}

func (a *ApproxCheck) startConsume() {
	var hardErr error
	for hardErr == nil {
		var msgBytes []byte
		msgBytes, hardErr = a.input.ReadBytes('\n')
		if hardErr != nil {
			break
		}
		hardErr = a.output.WriteBytes(msgBytes)
	}

	if hardErr == io.EOF {
		a.errMsg.LogFatal(nil, "Unexpected EOL listening for response input")
	} else {
		a.errMsg.LogFatal(nil, "Unexpected error listening for response input: %v", hardErr.Error())
	}
}

func (a *ApproxCheck) duration() time.Duration {
	switch a.speed {
	case SpeedFast:
		return 10 * time.Millisecond
	case SpeedModerate:
		return 200 * time.Millisecond
	default:
		return time.Second
	}
}

func (a *ApproxCheck) nextDateAction() *axmsg.Action {
	a.idCounter++
	r := axmsg.NewAction(
		nil,
		&a.idCounter,
		"date",
		nil,
		Date{
			Day:     a.date.Day(),
			Month:   int(a.date.Month()),
			Year:    a.date.Year(),
			Weekday: a.date.Weekday().String(),
		},
	)
	a.date = a.date.AddDate(0, 0, 1)
	return r
}

// Date ...
type Date struct {
	Day     int    `json:"day"`
	Month   int    `json:"month"`
	Year    int    `json:"year"`
	Weekday string `json:"weekday"`
}
