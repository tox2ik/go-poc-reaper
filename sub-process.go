package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Opt struct {
	Uid     int
	LogPath string
	Binary  string
	Args    []string
}

func opts(input []string) (Opt, []string) {

	uid := 1512
	logPath := fmt.Sprintf("/tmp/logfile.%d", uid)
	binary := "/tmp/zleep"
	args := []string{}

	if len(input) == 1 {
		if input[0] == "-h" || input[0] == "--help" {
			println(fmt.Sprintf(
				"Usage: %s [log-path] [uid] [binary-path] [binary-arguments]",
				path.Base(os.Args[0])))
			os.Exit(1)
		}
	}

	if len(input) >= 1 {
		logPath = input[0]
		input = input[1:]
	}
	if len(input) >= 1 {
		uid, _ = strconv.Atoi(input[0])
		input = input[1:]
	}
	if len(input) >= 1 {
		binary = input[0]
		input = input[1:]
	}

	tail := []string{}
	if len(input) >= 1 {
		sep := len(input) - 1
		if sep < 0 {
			sep = 0
		}
		last := ""
		for i, e := range input {
			last = input[i]
			if e == "--" {
				sep = i
				break
			}
		}
		if last == "--" {
			args = input[0:sep]
			tail = input[sep+1:]
		} else {
			args = input
		}

	}

	return Opt{
		Uid:     uid,
		LogPath: logPath,
		Binary:  binary,
		Args:    args,
	}, tail
}

func getOpts() []Opt {
	var rtt []Opt
	var o Opt

	aa := os.Args[1:]
	for len(aa) > 0 {
		o, aa = opts(aa)
		rtt = append(rtt, o)
	}
	return rtt
}

func info(format string, v ...interface{}) {
	println(fmt.Sprintf(format, v...))
}

func dump(rtt []Opt) {
	for _, rt := range rtt {
		s := fmt.Sprintf("%#v", rt)
		s = strings.ReplaceAll(s, " ", "\n   ")
		fmt.Printf("%s\n", s)
	}

}

func openLog(rt Opt) *os.File {
	tmpLog, errf := os.OpenFile(rt.LogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if errf != nil {
		info("can not log to %s: %s\n%#v", rt.LogPath, errf, errf)
		os.Exit(2)
	}
	errch := os.Chown(rt.LogPath, rt.Uid, rt.Uid)
	if errch != nil {
		info("chown %#v", errch)
	}
	return tmpLog
}

func ifEnv(s string) bool {
	return len(os.Getenv(s)) > 0
}

func main() {

	rtt := getOpts()

	donec := make(chan bool)
	wa := len(rtt)
	wg := sync.WaitGroup{}
	wg.Add(wa)

	for _, rt := range rtt {

		pid, errSub := Daemonish(rt.Binary, rt.Args, rt.Uid, openLog(rt), &donec)
		if errSub != nil {
			info("main() ERROR:\n%s", errSub)
			os.Exit(3)
		} else {
			info("main() Daemon away! %d (%s)", pid, rt.Binary)
		}
	}

	if ifEnv("CRASH") {
		info("main() CRASH imminent")
		println(exec.Command("foo").Process.Pid) // null pointer
	}

	if ifEnv("ABORT") {
		// don't wait for subprocess after all
		go func(subprocs int) {
			n, _ := strconv.Atoi(os.Getenv("ABORT"))
			if n > 0 {
				info("main() ABORT in %d", n)
				time.Sleep(time.Second * time.Duration(n))
				for subprocs > 0 {
					wg.Done()
					subprocs--
					time.Sleep(time.Millisecond)
				}
			}
		}(wa)
	}

	go func() {
		for {
			// time.Sleep(time.Millisecond * 777)
			info("main(1) select on done channel")
			select {
			case ok := <-donec:
				wg.Done()
				info("main(1) select on done channel: %t", ok)
			}
		}
	}()

	info("main() waiting for children")
	wg.Wait()
	info("main() done.")

}

func Daemonish(bin string, args []string, uid int, log *os.File, c *chan bool) (int, error) {
	cmd := exec.Command(bin, args...)
	cmd.Stdout = log

	if os.Getuid() == 0 {
		info("subp() setuid %d", uid)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Gid: uint32(uid),
				Uid: uint32(uid)}}
	} else {
		if uid != os.Getuid() {
			info("must be root to change process uid.")
		}
	}

	err := cmd.Start()
	if err != nil {
		// fmt.Printf("PERM %t\n", errors.Is(err, os.ErrPermission))
		return -1, fmt.Errorf(
			"command failed: %s\nFAILURE %#v\n%s",
			cmd.String(), err, err)
	}

	go func() {
		info("subp() cmd.Wait() [%s]", cmd.String())
		err := cmd.Wait() // Wait is necessary so cmd doesn't become a zombie
		info("subp() cmd.Wait() [%s] DONE", bin)
		*c <- true
		if err != nil {
			fmt.Printf("go func() Wait err: %s\n", err)
		}
	}()
	return cmd.Process.Pid, nil
}
