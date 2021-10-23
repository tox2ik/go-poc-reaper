https://stackoverflow.com/questions/59410139/run-command-in-golang-and-detach-it-from-process
----------


*Your question is imprecise or you are asking for non-standard features;*

>  In fact I want my background process to be as separate as possible: it has different parent pid, group pid etc. from my program. I want to run it as daemon.

*That's not how process inheritance works. You can not have process A start Process B and somehow change the parent pid of B to C. To the best of my knowledge this is not possible in Linux.*

*In other words, if process A (pid 55) starts process B (100), then B must have pid 55 as parent.*

*The only way to avoid that is have _something else_ start the B process such as atd, crond, or something else - which is not what you are asking for.*

> I want to run it as daemon.

*That's excellent. However, in a GNU / Linux system, all daemons have a parent pid and those parents have a parent pid going all the way up to pid 1, strictly according to the parent -> child rule.* 

---

With that out of the way...

---



So far you have already solved the following:

- redirect stdout of sub-process to a file
- set owner UID of sub-process
- sub-process survives death of parent (my program exits)
- the PID of sub-process can be seen by the main program

The last one is hard to reproduce, but I have heard about this problem in the docker world, so the answer is focused around addressing this issue.

- sub-process survives the parent being killed
- sub-process survives if the parent crashes (does not become a zombie)


---

### Case 5

    STAT COMMAND          USER     ELAPSED PID   PPID
    S    sh               0         0:01       1     0
    S    zleep            3         0:01      16     1
    Z    fork-if          3         0:01      17     1
    R    fork-child-A     3         0:01      18     1
    R    fork-child-B     3         0:01      21    18
    S    fork-child-C     3         0:01      25    21
    S    fork-daemon      3         0:01      26    25
    R    ps               0         0:01      27     1
    STAT COMMAND          USER     ELAPSED PID   PPID
    S    sh               0         0:02       1     0
    S    zleep            3         0:02      16     1
    Z    fork-if          3         0:02      17     1
    Z    fork-child-A     3         0:02      18     1
    R    fork-child-B     3         0:02      21     1
    S    fork-child-C     3         0:02      25    21
    S    fork-daemon      3         0:02      26    25
    R    ps               0         0:01      28     1
    STAT COMMAND          USER     ELAPSED PID   PPID
    S    sh               0         0:03       1     0
    S    zleep            3         0:03      16     1
    Z    fork-if          3         0:03      17     1
    Z    fork-child-A     3         0:03      18     1
    R    fork-child-B     3         0:03      21     1
    S    fork-child-C     3         0:03      25    21
    S    fork-daemon      3         0:03      26    25
    R    ps               0         0:01      29     1
    STAT COMMAND          USER     ELAPSED PID   PPID
    S    sh               0         0:03       1     0
    Z    zleep            3         0:03      16     1
    Z    fork-if          3         0:03      17     1
    Z    fork-child-A     3         0:03      18     1
    R    fork-child-B     3         0:03      21     1
    S    fork-child-C     3         0:03      25    21
    S    fork-daemon      3         0:03      26    25
    R    ps               0         0:01      30     1


### Demo


One terminal:
    make tail-cases

Another terminal

    make case0
    ...

Foo
    make sub-process
    make log

---


Related questions:

- https://stackoverflow.com/questions/23736046/how-to-create-a-daemon-process-in-golang
- https://stackoverflow.com/questions/10067295/how-to-start-a-go-program-as-a-daemon-in-ubuntu

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
