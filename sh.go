package main

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/shlex"
	"github.com/tox2ik/go-poc-reaper/fn"
)

func main() {

	args := os.Args[1:]
	script := args[0:0]
	if len(args) >= 1 {
		if args[0] == "-c" {
			script = args[1:]
		}
	}
	if len(script) == 0 {
		fn.CyanBold("cmd: expecting sh -c 'foobar'")
		os.Exit(111)
	}

	var cmd *exec.Cmd
	parts, _ := shlex.Split(strings.Join(script, " "))
	if len(parts) >= 2 {
		cmd = fn.Merge(exec.Command(parts[0], parts[1:]...), nil)
	}
	if len(parts) == 1 {
		cmd = fn.Merge(exec.Command(parts[0]), nil)
	}


	if fn.IfEnv("HANG") {
		fn.CyanBold("cmd: %v\n      start", parts)
		ex := cmd.Start()
		if ex != nil {
			fn.CyanBold("cmd %v err: %s", parts, ex)
		}
		go func() {
			time.Sleep(time.Millisecond * 100)
			errw := cmd.Wait()
			if errw != nil {
				fn.CyanBold("cmd %v err: %s", parts, errw)
			} else {
				fn.CyanBold("cmd %v all done.", parts)
			}
		}()

		fn.CyanBold("cmd: %v\n      dispatched, hanging forever (i.e. to keep docker running)", parts)
		for {
			time.Sleep(time.Millisecond * time.Duration(fn.EnvInt("HANG", 2888)))
			fn.SystemCyan("/bin/ps", "-e", "-o", "stat,comm,user,etime,pid,ppid")
		}


	} else {

		if fn.IfEnv("NOWAIT") {
			ex := cmd.Start()
			if ex != nil {
				fn.CyanBold("cmd %v start err: %s", parts, ex)
			}
		} else {

			ex := cmd.Run()
			if ex != nil {
				fn.CyanBold("cmd %v run err: %s", parts, ex)
			}
		}
		fn.CyanBold("cmd %v\n      dispatched, exit docker.", parts)
	}

	//time.Sleep(time.Millisecond * 100)
}
