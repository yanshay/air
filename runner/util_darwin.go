package runner

import (
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func (e *Engine) killCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid

	if e.config.Build.SendInterrupt {
		// Sending a signal to make it clear to the process that it is time to turn off
		if err = syscall.Kill(-pid, syscall.SIGINT); err != nil {
			return
		}
		time.Sleep(e.config.Build.KillDelay)
	}
	// https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
	err = syscall.Kill(-pid, syscall.SIGKILL)
	// Wait releases any resources associated with the Process.
	_, _ = cmd.Process.Wait()
	return pid, err
}

func (e *Engine) startCmd(cmd string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	c := exec.Command("/bin/sh", "-c", cmd)

	// When running Terminal UI Applications, not sure why, but when using the setpgid (which runs the process under a new pgid)
	// the application doesn't show. This change has implications on how process needs to be killed later on
	if !e.config.Build.TermUIApp {
		// because using pty cannot have same pgid
		c.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err = c.Start()
	if err != nil {
		return nil, nil, nil, err
	}
	return c, stdout, stderr, nil
}
