# Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
#
# Created by M. Massenzio, 2022-06-23

bin := out/bin/opatest

all:
	cd go && go build -o ../$(bin) cmd/main.go
