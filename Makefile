# Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
#
# Created by M. Massenzio, 2022-06-23

bin := out/bin/opatest
version := 0.1.0

all:
	go build -ldflags "-X main.Release=$(version) -X main.ProgName=$(shell basename $(bin))" \
		-o $(bin) cmd/main.go
