# Copyright (c) 2022 CopilotIQ Inc.  All rights reserved
#
# Created by M. Massenzio, 2022-06-23

gentests := out/bin/gentests

all:
	cd go && go build -o ../$(gentests) cmd/main.go
