
ps = busybox ps
awk = busybox awk
grep = busybox grep
mkdir = busybox mkdir
timeo = busybox timeout
trunk = busybox truncate -s0


main: prep build case6
test: prep case0 case1 case2 case3 case4 case5 case6
build: simple zleep sub fork ctrl-c zignal image
build2: reaper zleep sub fork ctrl-c zignal image
image:      ; docker build -q -t sp .
fork:   bin ; gcc fork-if.c -o bin/fork-if -static -std=gnu99 -pthread -D_GNU_SOURCE
simple: bin ; go build -o bin/sh           sh.go
reaper: bin ; go build -o bin/sh           reaper.go
zleep:  bin ; go build -o bin/zleep        sleep.go
ctrl-c: bin ; go build -o bin/ctrl-c       ctrl-c.go
zignal: bin ; go build -o bin/zignal       signal.go
sub:    bin ; go build -o bin/sub-process  sub-process.go
bin: ; $(mkdir) -p bin


uid = $(shell id -u)
here = $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
vol = -v $(here)/log:/log
drun = docker run -it --rm $(vol)
drunbg = docker run -i --rm $(vol)

help: ; go run sub-process.go -h || true
_pszom = $(ps) -eo stat,comm,user,etime,pid,ppid | $(grep) -e ^Z -e ^STAT
_psfor = $(ps) -eo stat,comm,user,etime,pid,ppid | $(grep) -e fork[-] | cat -n
pz: ; @$(_pszom)
pf: ; @$(_psfor)
pw: ; busybox watch -n 0.3 -t '$(_pszom); $(_psfor)'

tail-case: ; while true; do find log/ -name case\* | xargs timeout 40s xtail; done

prep: build perms
perms: export whoami=$(shell whoami)
perms: ; sudo make fixperms me=$${whoami}
fixperms:
	$(mkdir) -p log
	$(trunk) log/o || true
	for i in `busybox grep -E -e '^case[0-9]+' -o Makefile`;\
		do $(trunk) log/$${i} || true ; echo log/$${i}; done
	setfacl -m u:$(me):rwx log/o log/case* || true

# Just run sub-process with a short-lived child
case0:
	$(trunk) log/$@
	$(drun) $(dimg) /bin/sub-process /log/$@ 3 /bin/zleep 444 1
	$(grep) log/$@ -e. -c  | $(grep) -q ^4$$
	$(grep) log/$@ -e ^done

# Abort sub-process before children are done
case1:
	$(trunk) log/$@
	$(drun) -e ABORT=1 $(dimg) /bin/sub-process /log/$@ 3 /bin/zleep 444 10
	$(grep) -e. -c log/$@ | $(grep) -q ^3$$
	if $(grep) ^done log/$@ ; then exit 1; fi

# Abort after children are done (exit naturally)
case2:
	$(trunk) log/$@
	$(drun) -e ABORT=3 $(dimg) /bin/sub-process /log/$@ 3 /bin/zleep 444 1
	$(grep) -e. -c log/$@ | $(grep) -q ^4$$
	$(grep) ^done log/$@

# simulate bug in sub-process. docker will exit and stop the child prematurely
case3:
	$(trunk) log/$@
	$(drun) -e CRASH=yes $(dimg) /bin/sub-process /log/$@ 3 /bin/zleep 100 10 || true
	$(grep) -e. -c log/$@ | $(grep) -q ^1$$
	if $(grep) ^done log/$@ ; then exit 1; fi

# simulate bug in sub-process. sub-process will crash, but docker will remain running
# this should generate a zombie
case4:
	$(trunk) log/$@
	$(timeo) 3s $(drun) -e CRASH=yes -e HANG=1600 $(dimg) /bin/sh -c '/bin/sub-process /log/case4 3 /bin/zleep 222 1' || true
	$(grep) -e. -c log/case4 | $(grep) -q ^6$$
	$(grep) ^done log/case4


# Crash sub-process - children should survive, there will be zombies.
case5:
	$(trunk) log/$@
	$(timeo) 3s $(drun) -e CRASH=yes -e HANG=600 $(dimg) /bin/sh -c  '/bin/sub-process /log/$@ 3 /bin/zleep 111 2 -- /dev/stderr 3 /bin/fork-if --' || true
	$(grep) -e. -c log/$@ | $(grep) -q ^19$$
	$(grep) ^done log/$@


# Kill sub-process before it exits, children should survive
case6:
	$(trunk) log/$@
	@{ sleep 2; \
	  container=$$(docker container ls -a \
	      --format '{{ .State }} {{ .Image }} {{ .Names }} ' \
		  | $(awk) '/^running $(dimg)/{ print $$3 }'); \
	  docker exec -i -e NOWAIT=1 -e HANG='' $${container} sh -c 'killall sub-process'; \
	} &
	$(timeo) 4s \
      $(drun) \
      -e HANG=1200 $(dimg) /bin/sh -c  \
      '/bin/sub-process /log/$@ 3 /bin/zleep 888 10 -- /dev/stderr 3 /bin/fork-if --'

# Demonstrates how signals travel from parent to child
signs = SIGTERM SIGTERM SIGINT SIGHUP
case7:
	{ let i=5; \
      while [[ i -gt 0 ]]; do \
        let i=i-1; \
	    sleep 1; \
	  for s in $(signs); do \
		  pidof ctrl-c | xargs -r kill -$$s ;  \
	  done; done; \
	  sleep 3; killall -s KILL ctrl-c ; killall -s KILL zignal ; \
	} &
	@echo
	./bin/ctrl-c

case8:
	{ sleep 6 && killall -s KILL zignal; } &
	./bin/ctrl-c &
	sleep 2; killall -s TERM ctrl-c
	sleep 1; killall -s TERM ctrl-c
	sleep 1; killall -s TERM ctrl-c
	sleep 1; killall -s KILL ctrl-c




dcon = spc
dimg = sp

create:
	docker container rm $(dcon) || true
	@docker create -h sph --name $(dcon) $(vol) sp /bin/ps > /dev/null
	@echo
	@docker export $(dcon) | tar tf - | sort -V


clean:
	rm -fv ./bin/sh ./bin/zleep ./bin/sub-process ./bin/fork-if
	docker container rm $(dcon) || true
	docker image rm $(dimg) || true


fork-stop:
	@killall fork-daemon || true
	@killall fork-child-A || true
	@killall fork-child-B || true
	@killall fork-child-C || true
	@killall -s KILL ctrl-c  || true
	@killall -s KILL zignal  || true
fork-restart: fork-stopstop
	make fork-if && ./fork-if
