#!/bin/sh
go build -ldflags "-X main.gh_repo blang/hornet -X main.webif_url http://webinterfaceurl -X main.version 0.0.1" .
