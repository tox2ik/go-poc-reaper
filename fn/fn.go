package fn

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/fatih/color"
)

var FgMagenta = color.New(color.FgMagenta)
var FgCyan = color.New(color.FgCyan)
var FgReset = color.New(color.Reset)

func System(comm string, arg ...string) {
	SystemX(FgReset, comm, arg...)
}

func SystemCyan(comm string, arg ...string) {
	SystemX(FgCyan, comm, arg...)
}

func SystemX(colore *color.Color, comm string, arg ...string) {
	cmd := Merge(exec.Command(comm, arg...), colore)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%#v", err)
	}
}


type cwrite struct {
	clr *color.Color
	sink io.Writer
}
func (w cwrite) Write(p []byte) (n int, err error) {
	return w.clr.Fprint(w.sink, string(p))
}

func Merge(comm *exec.Cmd, colore *color.Color) *exec.Cmd {
	if colore != nil {
		out := cwrite{
			clr: colore,
			sink: os.Stdout,
		}
		comm.Stdout = out
		comm.Stderr = out
	} else {
		comm.Stdout = os.Stdout
		comm.Stderr = os.Stdout
	}
	return comm
}

var cCyanBold = color.New(color.FgCyan, color.Bold)
var cMagenta = color.New(color.FgMagenta)

func CyanBold(format string, v ...interface{}) {
	println(cCyanBold.Sprintf(format, v...))
	color.Unset()
}

func Magenta(format string, v ...interface{}) {
	_,_ = cMagenta.Fprint(os.Stderr, fmt.Sprintf(format, v...) + "\n")
	color.Unset()
}


func IfEnv(s string) bool {
	return len(os.Getenv(s)) > 0
}

func EnvInt(key string, ifnone int) int {
	n, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return ifnone
	}
	return n
}
