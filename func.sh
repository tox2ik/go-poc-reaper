sp () { docker run -it --rm -v $(pwd)/log:/log sp "$@"; }
sph () { docker run -it --rm -v $(pwd)/log:/log -e HANG=300 sp "$@"; }

capture() {
	for i in {1..20}; do echo $i;
		import -window root ~/tmp/ss/test-a.$i.png; done
}


conname () {
	docker container ls -a --format '{{ .State }} {{ .Image }} {{ .Names }} ' \
		| busybox awk '/^running sp/{ print $3 }'
}


gorun_bg() { go run ctrl-c.go & cmd=$!; echo pid $cmd; }
