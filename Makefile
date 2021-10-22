.PHONY: fork-if

spp=$(shell cat /tmp/sub-proc 2>/dev/null || echo 1)
uid=$(shell id -u)
here = $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

vol=-v $(here)/log:/log

run: sleep
	sudo kill $(spp) || true
	sudo rm /tmp/sub-proc || true
	sudo go run sub-process.go /tmp/log $(uid) /tmp/zleep

pz: ; ps -eos,cmd | grep  ^Z
ps:
	ps -o stat,user,pid,cmd -p $(spp)
	@echo ==
	ps -eo s,stat,comm,cmd | grep ^Z
	@echo ==
	ps -e -o s,comm,pid,ppid | grep -e fork[-] | cat -n
help:  ; go run sub-process.go -h || true
log:
	ls -l /tmp/log
	cat -n /tmp/log

fork-if: ; gcc fork-if.c -o bin/fork-if -static -std=gnu99 -pthread -D_GNU_SOURCE
simple: ; go build -o bin/sh           sh.go
sleep:  ; go build -o bin/zleep        sleep.go
sub:    ; go build -o bin/sub-process  sub-process.go
build: sleep sub simple fork-if
	/bin/mkdir -p log
	/bin/mkdir -p bin
	touch log/.keep
	docker build -q -t sp .

create:
	docker container rm spc || true
	docker create -h sph --name spc $(vol) sp /bin/ps
	@echo
	@docker export spc | tar tf - | sort -V

tail-case: ; while true; do find log/ -name case\* | xargs timeout 40s xtail; done

# Just run sub-process with a short-lived child
case0:
	sudo truncate -s0 log/case0
	docker run --rm $(vol) sp /bin/sub-process /log/case0 3 /bin/zleep 444 1
	grep -e. -c log/case0 | grep -q ^4$$
	grep ^done log/case0

# Abort sub-process before children are done
case1:
	sudo truncate -s0 log/case1
	docker run --rm $(vol) -e ABORT=1 sp /bin/sub-process /log/case1 3 /bin/zleep 444 10
	grep -e. -c log/case1 | grep -q ^3$$
	if grep ^done log/case1 ; then exit 1; fi

# Abort after children are done (exit naturally)
case2:
	sudo truncate -s0 log/case2
	docker run --rm $(vol) -e ABORT=3 sp /bin/sub-process /log/case2 3 /bin/zleep 444 1
	grep -e. -c log/case2 | grep -q ^4$$
	grep ^done log/case2

# simulate bug in sub-process. docker will exit and stop the child prematurely
case3:
	sudo truncate -s0 log/case3
	docker run --rm $(vol) -e CRASH=yes sp /bin/sub-process /log/case3 3 /bin/zleep 100 10 || true
	grep -e. -c log/case3 | grep -q ^1$$
	if grep ^done log/case3 ; then exit 1; fi

# simulate bug in sub-process. sub-process will crash, but docker will remain running
# this should generate a zombie
case4:
	sudo truncate -s0 log/case4
	timeout 3s docker run --rm $(vol) -e CRASH=yes -e HANG=1600 sp /bin/sh -c \
                                                                '/bin/sub-process /log/case4 3 /bin/zleep 222 1' \
                                                                || true
	grep -e. -c log/case4 | grep -q ^6$$
	grep ^done log/case4


# Crash sub-process - sub commands should survive, there will be zombies.
case5:
	sudo truncate -s0 log/case5z
	sudo truncate -s0 log/case5f
	timeout 5s docker run --rm $(vol) -e CRASH=yes -e HANG=600 sp /bin/sh -c \
                                  '/bin/sub-process /log/case5z 3 /bin/zleep 111 2 -- /log/case5f 3 /bin/fork-if --' \
                                  || true
	grep -e. -c log/case5z | grep -q ^19$$
	grep ^done log/case5z


clean:
	rm -fv ./bin/sh ./bin/zleep ./bin/sub-process ./bin/fork-if
	setfacl -m u:$(shell whoami):rwx log/o
	truncate -s0 log/o;


fork-stop:
	@killall fork-daemon || true
	@killall fork-child-A || true
	@killall fork-child-B || true
	@killall fork-child-C || true
fork-restart: stop
	make fork-if && ./fork-if
fork-ps:
	watch -n0,1 'ps -e -o s,comm,pid,ppid | grep -e fork[-] | cat -n'
