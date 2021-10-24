# go-poc-reaper

Experiments with having a go binary as PID 1.

It also includes some experiments for launching binaries from go
in a shell environment or from make.

This was started as an answer on Stack Overflow but grew into a 
three day experiment because I wanted to understand how signals 
work for child processes and what zombies are and how to deal with them.


![Case5 - reaper](https://raw.githubusercontent.com/tox2ik/go-poc-reaper/main/.case5-reaper.webp)
