package main

import (
	"github.com/jaqmol/approx/axenvs"
	"github.com/jaqmol/approx/axmsg"
)

func main() {
	envs := axenvs.NewEnvs("approx_check", []string{"MODE"}, []string{"SPEED"})
	errMsg := axmsg.Errors{Source: "approx_check"}

	if len(envs.Outs) != 1 {
		errMsg.LogFatal(nil, "Check expects exactly 1 output, but got %v", len(envs.Outs))
	}
	if len(envs.Ins) != 1 {
		errMsg.LogFatal(nil, "Check expects more than 1 input, but got %v", len(envs.Ins))
	}

	if len(envs.Required["MODE"]) == 0 {
		errMsg.LogFatal(nil, "Check expects value for env MODE")
	}

	at := NewApproxCheck(envs)
	at.Start()
}
