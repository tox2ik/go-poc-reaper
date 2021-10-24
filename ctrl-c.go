package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/tox2ik/go-poc-reaper/fn"
)

func signs(s ...os.Signal) chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, s...)
	signal.Notify(signals,
		os.Interrupt, syscall.SIGINT, syscall.SIGQUIT, // keyboard
		syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM, // os termination
		syscall.SIGUSR1, syscall.SIGUSR2, // user
		syscall.SIGPIPE, syscall.SIGCHLD, syscall.SIGSEGV, // os other
	)
	return signals
}

func interpret(signals chan os.Signal) chan os.Signal {
	go func() {
		for {
			select {
			case sign := <-signals:
				elog("go main() got %#v %s", sign, sign)
			}
		}
	}()
	return signals
}

func bg(comm string, argz string) {
	cmd := exec.Command(comm, strings.Split(argz, " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	go func() {
		err := cmd.Run()
		if err != nil {
			elog("%v", err)
		}
	}()

	time.Sleep(time.Millisecond * 88)
	elog("zignal %d (%d)", cmd.Process.Pid, syscall.Gettid())

}

func bash(donec chan error, script string) {
	cmd := exec.Command("/bin/bash", "-c", script)
	cmd.Stdout = os.Stderr
	err := cmd.Start()

	go func() {
		err = cmd.Wait()
		donec <- err
	}()
	if err != nil {
		log.Fatal(err)
	}
	elog("bashed %d (%d)", cmd.Process.Pid, syscall.Gettid())
}

func main() {

	doit()

}

func doit() {
	println("kill with:\n\t{ pidof ctrl-c; pidof signal ; } | xargs -r -t kill  -9 ")
	fmt.Println("main()", os.Getpid())

	if ! fn.IfEnv("NOSIGN") {
		interpret(signs())
	}

	donec := make(chan error, 1)

	bash(donec, `
        trap ' echo Bash _ $$  INTs ignored; ' SIGINT
        trap ' echo Bash _ $$ QUITs ignored; ' SIGQUIT
        trap ' echo Bash _ $$ EXITs $( ps -o etimes -p $$ )' EXIT
        sleep 6;
    `)

	bash(donec, `
        trap ' echo Bash q $$  INTs ignored;   ' SIGINT
        trap ' echo Bash q $$ QUITs quit; exit ' SIGQUIT
        trap ' echo Bash q $$ EXITs $( ps -o etimes -p $$ )' EXIT
        sleep 8;
    `)

	bash(donec, `
        trap ' echo Bash c $$  INTs interrupt; exit ' SIGINT
        trap ' echo Bash c $$ QUITs ignored;        ' SIGQUIT
        trap ' echo Bash c $$ EXITs $( ps -o etimes -p $$ )'  EXIT
        sleep 4
    `)
	bg("bin/zignal", "int")
	bg("bin/zignal", "int quit")
	bg("bin/zignal", "int quit term")


	children := int32(6)
	i := int32(0)

	if fn.IfEnv("NOWAIT") {

		go func() {
			for {
				time.Sleep(time.Millisecond * 666)
				elog("main() {%d} %d", i, os.Getpid())
			}
		}()
		// wait for interactive ctrl-c or ctrl-\
		//time.Sleep(2 * time.Second)
		time.Sleep(6 * time.Second)
	} else {
	wait:
		for {
			elog("main() wait...")
			select {
			case err := <-donec:
				atomic.AddInt32(&i, 1)
				elog("main() children done: %d %s", i, err)
				if i == children {
					close(donec)
					break wait
				}
			}
		}
	}
	elog("main() done.")
}

func echo(a ...interface{}) {
	_, err := fmt.Println(a...)
	if err != nil {
		fmt.Println("ERR ", err.Error())
	}
}

func elog(form string, arg ...interface{}) {
	println(fmt.Sprintf(form, arg...))
}
