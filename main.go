package main

import (
	"github.com/jaqmol/approx/errormsg"
	"github.com/jaqmol/approx/processorconf"
)

func main() {
	conf := processorconf.NewProcessorConf("approx_test", []string{"MODE", "SPEED"})
	errMsg := errormsg.ErrorMsg{Processor: "approx_test"}

	if len(conf.Outputs) != 1 {
		errMsg.LogFatal(nil, "Test expects exactly 1 output, but got %v", len(conf.Outputs))
	}
	if len(conf.Inputs) != 1 {
		errMsg.LogFatal(nil, "Test expects more than 1 input, but got %v", len(conf.Inputs))
	}

	if len(conf.Envs["MODE"]) == 0 {
		errMsg.LogFatal(nil, "Test expects value for env MODE")
	}

	at := NewApproxCheck(conf)
	at.Start()
}
