FROM scratch

# for sh.go
ENV HANG ""

# for sub-process.go
ENV ABORT ""
ENV CRASH ""
ENV KILL ""

# for ctrl-c.go, signal.go
ENV NOSIGN ""

COPY bin/sh          /bin/sh
COPY bin/sub-process /bin/sub-process
COPY bin/zleep       /bin/zleep
COPY bin/fork-if     /bin/fork-if


COPY --from=busybox:latest /bin/find    /bin/find
COPY --from=busybox:latest /bin/ls      /bin/ls
COPY --from=busybox:latest /bin/ps      /bin/ps
COPY --from=busybox:latest /bin/killall /bin/killall
#COPY --from=busybox:latest /bin/sh /bin/ash

#ENTRYPOINT /bin/sub-process /log/o 3 /bin/zleep 444 10
#ENTRYPOINT /bin/sub-process /log/o 3 /bin/zleep 444 10 -- /log/o 3 /bin/fork-if --
#ENTRYPOINT /bin/find
