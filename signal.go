package main

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var Signals = [...]string{
	"SIG___UNPECIFIED",
	"SIGHUP",
	"SIGINT",
	"SIGQUIT",
	"SIGILL",
	"SIGTRAP",
	"SIGABRT",
	"SIGBUS",
	"SIGFPE",
	"SIGKILL",
	"SIGUSR1",
	"SIGSEGV",
	"SIGUSR2",
	"SIGPIPE",
	"SIGALRM",
	"SIGTERM",
	"SIGSTKFLT",
	"SIGCHLD",
	"SIGCONT",
	"SIGSTOP",
	"SIGTSTP",
	"SIGTTIN",
	"SIGTTOU",
	"SIGURG",
	"SIGXCPU",
	"SIGXFSZ",
	"SIGVTALRM",
	"SIGPROF",
	"SIGWINCH",
	"SIGIO",
	"SIGPWR",
	"SIGSYS",
}

func sig(input string) os.Signal {
	up := strings.ToUpper(input)
	up3 := strings.ToUpper(input)
	for sn, name := range Signals {
		if up == name || up3 == name[3:] || input == name {
			return syscall.Signal(sn)
		}
	}
	n := 0
	if i, err := strconv.Atoi(input); i > 0 && err == nil {
		n = i
	}
	return syscall.Signal(n)
}

func main() {

	notify := []os.Signal{}
	ignore := []os.Signal{}
	for _, input := range os.Args[1:] {
		if input[0:1] == "-" {
			s := sig(input[1:])
			if s != syscall.Signal(0) {
				ignore = append(ignore, s)
			}
		} else {
			s := sig(input)
			if s != syscall.Signal(0) {
				notify = append(notify, s)
			}
		}
	}

	ossig := make(chan os.Signal)

	if len(notify) > 0 {
		signal.Notify(ossig, notify...)
	}

	if len(ignore) > 0 {
		signal.Ignore(ignore...)
	}

	if len(notify) + len(ignore) == 0 {
		println("no signals specified (int, -term, -quit, alrm, pipe)")
		os.Exit(1)
	}

	go func() {
		for ;; {
			select {
			case s := <-ossig:
				println("pid ", os.Getpid(), s.String())
			}
		}
	}()

	for ;;  {
		time.Sleep(time.Second * 1)
		println("p ", os.Getpid())
	}

}
