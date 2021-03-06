package server

import (
	"errors"
	"io"
	"time"
)

// RunTask is a task (script) for a Run
type RunTask struct {
	ScriptName   string
	ScriptReader io.Reader
	As           string
}

// Run is a list of Tasks on Host, including task results
type Run struct {
	SSHConn *SSHConnection
	Tasks   []*RunTask
	// CurrentTask int
	// StartTime    time.Time
	// Duration     time.Duration
	// DialDuration time.Duration
	Log *Log
}

// Go will execute the Run
func (run *Run) Go() error {
	const bootstrap = "bash -s --"
	errorChan := make(chan error)

	if len(run.Tasks) == 0 {
		run.Log.Info("nothing to run")
		return nil
	}

	if err := run.SSHConn.Connect(); err != nil {
		return err
	}
	defer run.SSHConn.Close()

	if err := run.preparePipes(errorChan); err != nil {
		return err
	}

	if err := run.SSHConn.Session.Run(bootstrap); err != nil {
		return err
	}

	// currently, I'm not sure that I will have and errorChan
	// in every case, so… let's timeout.
	select {
	case err := <-errorChan:
		return err
	case <-time.After(1 * time.Second):
		return errors.New("timeout after waiting stderr errorChan")
	}
}
