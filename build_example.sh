#!/bin/sh
go build -ldflags "-X main.pushr_host http://pushr.example.com  -X main.pushr_release hornet -X main.pushr_channel stable -X main.pushr_readtoken $PUSHR_READ -X main.webif_url http://webinterfaceurl -X main.version 0.0.1" .
