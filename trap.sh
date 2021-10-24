#!/bin/bash

trap ' echo INTs; ' SIGINT
trap ' echo QUITs; ' SIGQUIT
trap ' echo EXITs; ' EXIT
sleep 1
trap - SIGQUIT
