package main

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/jaqmol/approx/axenvs"
	"github.com/jaqmol/approx/axmsg"
)

// NewApproxCheck ...
func NewApproxCheck(envs *axenvs.Envs) *ApproxCheck {
	errMsg := &axmsg.Errors{Source: "approx_check"}

	modeEnv := envs.Required["MODE"]
	var mode Mode
	switch modeEnv {
	case "check":
		mode = ModeCheck
	case "collect":
		mode = ModeCollect
	case "tick":
		mode = ModeTick
	default:
		errMsg.LogFatal(nil, "Check expects env MODE to be either check, collect or tick but got %v", modeEnv)
	}

	var expect []string
	if ModeCheck == mode {
		expectEnv, ok := envs.Optional["EXPECT"]
		if !ok {
			errMsg.LogFatal(nil, "Check expects env EXPECT, if MODE is check")
		}
		rawExpect := strings.Split(expectEnv, ",")
		expect = make([]string, 0)
		for _, rawValue := range rawExpect {
			value := strings.TrimSpace(rawValue)
			expect = append(expect, value)
		}
	}

	var speed Speed
	if ModeTick == mode {
		speedEnv, ok := envs.Optional["SPEED"]
		if !ok {
			errMsg.LogFatal(nil, "Check expects env SPEED, if MODE is tick")
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
			errMsg.LogFatal(nil, "Check expects env SPEED to be either untethered, fast, moderate or slow, but got %v", modeEnv)
		}
	}

	ins, outs := envs.InsOuts()

	return &ApproxCheck{
		errMsg:  errMsg,
		output:  axmsg.NewWriter(&outs[0]),
		input:   axmsg.NewReader(&ins[0]),
		mode:    mode,
		expect:  expect,
		speed:   speed,
		counter: 0,
		date:    time.Now(),
	}
}

// ApproxCheck ...
type ApproxCheck struct {
	errMsg  *axmsg.Errors
	output  *axmsg.Writer
	input   *axmsg.Reader
	mode    Mode
	expect  []string
	speed   Speed
	counter int
	date    time.Time
}

// Mode ...
type Mode int

// Mode Types
const (
	ModeCheck Mode = iota
	ModeCollect
	ModeTick
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
	if ModeCheck == a.mode {
		a.startCheck()
	} else if ModeCollect == a.mode {
		a.startCollect()
	} else if ModeTick == a.mode {
		if SpeedUntethered == a.speed {
			a.startUntetheredTick()
		} else {
			a.startTetheredTick()
		}
	}
}

func (a *ApproxCheck) startCheck() {
	msg := a.checkAction()
	err := a.output.Write(msg)
	if err != nil {
		a.errMsg.Log(&a.counter, "Error writing request message to output: %v", err.Error())
		return
	}

	actn, rawData, err := a.input.Read()
	if actn.AXMSG == 1 && actn.Role == "check" {
		var data Check
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			a.errMsg.LogFatal(actn.ID, "Error parsing response data: %v", err.Error())
		}

		procsNotFound := make([]string, 0)
		for _, expName := range a.expect {
			didFind := false
			for _, prcName := range data.Processors {
				if expName == prcName {
					didFind = true
					break
				}
			}
			if !didFind {
				procsNotFound = append(procsNotFound, expName)
			}
		}

		if len(data.Processors) > 0 && len(procsNotFound) == 0 {
			a.counter++
			s := axmsg.NewAction(
				&a.counter,
				nil,
				"check-success",
				nil,
				Success{Success: true},
			)
			a.output.Write(s)
			if err != nil {
				a.errMsg.LogFatal(actn.ID, "Error writing success message: %v", err.Error())
			}
		} else {
			a.errMsg.LogFatal(actn.ID, "Check failed, processors %v not found", strings.Join(procsNotFound, ", "))
		}
	}
}

func (a *ApproxCheck) startCollect() {
	var hardErr error
	for hardErr == nil {
		var msgBytes []byte
		msgBytes, hardErr = a.input.ReadBytes()
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

func (a *ApproxCheck) startUntetheredTick() {
	for {
		a.nextTick()
	}
}

func (a *ApproxCheck) startTetheredTick() {
	ticker := time.NewTicker(a.duration())
	for range ticker.C {
		a.nextTick()
	}
}

func (a *ApproxCheck) nextTick() {
	msg := a.nextDateAction()
	err := a.output.Write(msg)
	if err != nil {
		a.errMsg.Log(&a.counter, "Error writing request message to output: %v", err.Error())
		return
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

func (a *ApproxCheck) checkAction() *axmsg.Action {
	a.counter++
	cmd := "add-processor-name"
	r := axmsg.NewAction(
		&a.counter,
		nil,
		"check",
		&cmd,
		Check{
			Processors: make([]string, 0),
		},
	)
	a.date = a.date.AddDate(0, 0, 1)
	return r
}

func (a *ApproxCheck) nextDateAction() *axmsg.Action {
	a.counter++
	r := axmsg.NewAction(
		nil,
		&a.counter,
		"tick",
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

// Check ...
type Check struct {
	Processors []string `json:"processors"`
}

// Success ...
type Success struct {
	Success bool `json:"success"`
}
