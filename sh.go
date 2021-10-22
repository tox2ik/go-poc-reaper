package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"
)

func merge(comm *exec.Cmd) *exec.Cmd {
	comm.Stdout = os.Stdout
	comm.Stderr = os.Stdout
	return comm
}

func main() {
	args := os.Args[1:]
	script := args[0:0]
	if len(args) >= 1 {
		if args[0] == "-c" {
			script = args[1:]
		}
	}

	if len(script) == 0 {
		fmt.Println("cmd: was expexting sh -c 'foobar'")
		os.Exit(111)
	}

	parts, _ := shlex.Split(strings.Join(script, " "))

	out := []byte{}
	var ex error

	if len(parts) >= 2 {

		fmt.Printf("cmd...: %v\n", parts)
		ex = merge(exec.Command(parts[0], parts[1:]...)).Run()
	}

	if len(parts) == 1 {
		fmt.Printf("cmd1: %v\n", parts)
		ex = merge(exec.Command(parts[0])).Run()
	}

	if len(out) > 0 {
		fmt.Println("cmd out:" + string(out))
	}

	if ex != nil {
		fmt.Printf("cmd err: %s\n", ex)
	}

	if len(os.Getenv("HANG")) > 0 {
		n, _ := strconv.Atoi(os.Getenv("HANG"))
		if n == 0 {
			n = 2888
		}
		fmt.Println("cmd: dispatched, hanging forever (i.e. to keep docker running)")
		for {
			system("/bin/ps", "-e", "-o", "stat,comm,user,etime,pid,ppid")
			time.Sleep(time.Millisecond * time.Duration(n))
			// o, _ := exec.Command().Output()
			// if len(o) > 0 {
			//	fmt.Printf(string(o))
			// }

		}
	} else {
		fmt.Printf("cmd dispatched, exit docker.\n")
	}

}

func system(comm string, arg ...string) {
	cmd := merge(exec.Command(comm, arg...))
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%#v", err)
	}
}
