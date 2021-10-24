https://stackoverflow.com/questions/59410139/run-command-in-golang-and-detach-it-from-process
----------


*Your question is imprecise or you are asking for non-standard features. *

> In fact I want my background process to be as separate as possible: it has different parent pid, group pid etc. from my program. I want to run it as daemon.

That is not how process inheritance works. You can not have process A start Process B and somehow change the parent of B to C. To the best of my knowledge this is not possible in Linux.

In other words, if process A (pid 55) starts process B (100), then B must have parent pid 55.

The only way to avoid that is have _something else_ start the B process such as atd, crond, or something else - which is not what you are asking for.

If parent 55 dies, then PID 1 will be the parent of 100, not some arbitrary process.

Your statement "it has different parent pid" does not makes sense.

> I want to run it as daemon.

That's excellent. However, in a GNU / Linux system, all daemons have a parent pid and those parents have a parent pid going all the way up to pid 1, strictly according to the parent -> child rule.

> when I send SIGTERM/SIGKILL to my program the underlying process crashes. 

I can not reproduce that behavior. See case8 and case7 from the proof-of-concept repo.

    make case8
    export NOSIGN=1; make build case7 
    unset NOSIGN; make build case7

In both cases the sub-processes continue to run despite main process being signaled with TERM, HUP, INT. This behavior is different in a shell environment because of convenience reasons. See the related questions about signals. [This particular answer](https://unix.stackexchange.com/a/387008/23421) illustrates a key difference for SIGINT. Note that SIGSTOP and SIGKILL cannot be caught by an application.

---

It was necessary to clarify the above before proceeding with the other parts of the question.


So far you have already solved the following:

- redirect stdout of sub-process to a file
- set owner UID of sub-process
- sub-process survives death of parent (my program exits)
- the PID of sub-process can be seen by the main program

The next one depends on whether the children are "attached" to a shell or not

- sub-process survives the parent being killed

The last one is hard to reproduce, but I have heard about this problem in the docker world, so the rest of this answer is focused on addressing this issue.

- sub-process survives if the parent crashes and does not become a zombie


As you have noted, the `Cmd.Wait()` is necessary to avoid creating zombies. After some experimentation I was able to consistency produce zombies ina a docker environment using an intentionally simple replacement for `/bin/sh`. This "shell" implemented in go will only run a single command and not much else in terms of reaping children. You can study the code [over at github](https://github.com/tox2ik/go-poc-reaper).


### Case 5 (simple /bin/sh)

The gist of it is we start two sub-processes from go, using the "parent" `sub-process` binary. The first child is `zleep` and the second `fork-if`. The second one starts a "daemon" that runs a forever-loop in addition to a few short-lived threads. After a while, we kill the `sub-procss` parent, forcing `sh` to take over the parenting for these children.

Since this simple implementation of sh does not know how to deal with abandoned children, the children become zombies.
This is standard behavior. To avoid this, [init systems are usually responsible](https://github.com/krallin/tini) for cleaning up any such children.

Check out this repo and run the cases:


    $ make prep build
    $ make prep build2

The first one will use the simple /bin/sh in the docker container, and the socond one will use the same code wrapped in a reaper.

With zombies:

    $ make prep build case5
    (…)
    main() Daemon away! 16 (/bin/zleep)
    main() Daemon away! 22 (/bin/fork-if)
    (…)
    main() CRASH imminent
    panic: runtime error: invalid memory address or nil pointer dereference
    [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x49e45c]
    goroutine 1 [running]:
    main.main()
        /home/jaroslav/src/my/go/doodles/sub-process/sub-process.go:137 +0xfc
    cmd [/bin/sub-process /log/case5 3 /bin/zleep 111 2 -- /dev/stderr 3 /bin/fork-if --] err: exit status 2
    Child '1' done
    thread done
    STAT COMMAND          USER     ELAPSED PID   PPID
    R    sh               0         0:02       1     0
    S    zleep            3         0:02      16     1
    Z    fork-if          3         0:02      22     1
    R    fork-child-A     3         0:02      25     1
    R    fork-child-B     3         0:02      26    25
    S    fork-child-C     3         0:02      27    26
    S    fork-daemon      3         0:02      28    27
    R    ps               0         0:01      30     1
    Child '2' done
    thread done
    daemon
    (…)
    STAT COMMAND          USER     ELAPSED PID   PPID
    R    sh               0         0:04       1     0
    Z    zleep            3         0:04      16     1
    Z    fork-if          3         0:04      22     1
    Z    fork-child-A     3         0:04      25     1
    R    fork-child-B     3         0:04      26     1
    S    fork-child-C     3         0:04      27    26
    S    fork-daemon      3         0:04      28    27
    R    ps               0         0:01      33     1
    (…)

With reaper:

    $ make -C ~/src/my/go/doodles/sub-process case5
    (…)
    main() CRASH imminent
    (…)
    Child '1' done
    thread done
    raeper pid 24
    STAT COMMAND          USER     ELAPSED PID   PPID
    S    sh               0         0:02       1     0
    S    zleep            3         0:01      18     1
    R    fork-child-A     3         0:01      27     1
    R    fork-child-B     3         0:01      28    27
    S    fork-child-C     3         0:01      30    28
    S    fork-daemon      3         0:01      31    30
    R    ps               0         0:01      32     1
    Child '2' done
    thread done
    raeper pid 27
    daemon
    STAT COMMAND          USER     ELAPSED PID   PPID
    S    sh               0         0:03       1     0
    S    zleep            3         0:02      18     1
    R    fork-child-B     3         0:02      28     1
    S    fork-child-C     3         0:02      30    28
    S    fork-daemon      3         0:02      31    30
    R    ps               0         0:01      33     1
    STAT COMMAND          USER     ELAPSED PID   PPID
    S    sh               0         0:03       1     0
    S    zleep            3         0:02      18     1
    R    fork-child-B     3         0:02      28     1
    S    fork-child-C     3         0:02      30    28
    S    fork-daemon      3         0:02      31    30
    R    ps               0         0:01      34     1
    raeper pid 18
    daemon
    STAT COMMAND          USER     ELAPSED PID   PPID
    S    sh               0         0:04       1     0
    R    fork-child-B     3         0:03      28     1
    S    fork-child-C     3         0:03      30    28
    S    fork-daemon      3         0:03      31    30
    R    ps               0         0:01      35     1
    (…)


Here is a picture of the same output, which may be less confusing to read.


Zombies

![Case5 - zombies](https://raw.githubusercontent.com/tox2ik/go-poc-reaper/main/.case5-simple.webp)

Reaper

![Case5 - reaper](https://raw.githubusercontent.com/tox2ik/go-poc-reaper/main/.case5-reaper.webp)

### Case5 (reaper /bin/sh)



### How to run the cases in the poc repo

Get the code

    git clone ...


One terminal:

    make tail-cases

Another terminal

    make build
    or make build2
    make case0 case1
    ...


---


Related questions:

go

- https://stackoverflow.com/questions/23736046/how-to-create-a-daemon-process-in-golang
- https://stackoverflow.com/questions/10067295/how-to-start-a-go-program-as-a-daemon-in-ubuntu
- https://stackoverflow.com/questions/42471349/how-to-keep-subprocess-running-after-program-exit-in-golang
- https://stackoverflow.com/questions/33165530/prevent-ctrlc-from-interrupting-exec-command-in-golang

signals

- https://unix.stackexchange.com/questions/386999/what-terminal-related-signals-are-sent-to-the-child-processes-of-the-shell-direc
- https://unix.stackexchange.com/questions/6332/what-causes-various-signals-to-be-sent
- https://en.wikipedia.org/wiki/Signal_(IPC)#List_of_signals


Related discussions:

- https://github.com/golang/go/issues/227
- https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/

Relevant projects:

- http://software.clapper.org/daemonize/ (what I would use)
- https://github.com/hashicorp/go-reap (if you must have run go on pid 1)
- https://github.com/sevlyar/go-daemon (mimics posix fork)


Relevant prose:

> A zombie process is a process whose execution is completed but it still has an entry in the process table. Zombie processes usually occur for child processes, as the parent process still needs to read its child’s exit status. Once this is done using the wait system call, the zombie process is eliminated from the process table. This is known as reaping the zombie process.

from [https://www.tutorialspoint.com/what-is-zombie-process-in-linux](tutorialspoint.com/what-is-zombie-process-in-linux)

---
