// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// mksysnum_plan9.sh /opt/plan9/sys/src/libc/9syscall/sys.h
// MACHINE GENERATED BY THE ABOVE COMMAND; DO NOT EDIT

package plan9

const (
	SYS_SYSR1       = 0
	SYS_BIND        = 2
	SYS_CHDIR       = 3
	SYS_CLOSE       = 4
	SYS_DUP         = 5
	SYS_ALARM       = 6
	SYS_EXEC        = 7
	SYS_EXITS       = 8
	SYS_FAUTH       = 10
	SYS_SEGBRK      = 12
	SYS_OPEN        = 14
	SYS_OSEEK       = 16
	SYS_SLEEP       = 17
	SYS_RFORK       = 19
	SYS_PIPE        = 21
	SYS_CREATE      = 22
	SYS_FD2PATH     = 23
	SYS_BRK_        = 24
	SYS_REMOVE      = 25
	SYS_NOTIFY      = 28
	SYS_NOTED       = 29
	SYS_SEGATTACH   = 30
	SYS_SEGDETACH   = 31
	SYS_SEGFREE     = 32
	SYS_SEGFLUSH    = 33
	SYS_RENDEZVOUS  = 34
	SYS_UNMOUNT     = 35
	SYS_SEMACQUIRE  = 37
	SYS_SEMRELEASE  = 38
	SYS_SEEK        = 39
	SYS_FVERSION    = 40
	SYS_ERRSTR      = 41
	SYS_STAT        = 42
	SYS_FSTAT       = 43
	SYS_WSTAT       = 44
	SYS_FWSTAT      = 45
	SYS_MOUNT       = 46
	SYS_AWAIT       = 47
	SYS_PREAD       = 50
	SYS_PWRITE      = 51
	SYS_TSEMACQUIRE = 52
	SYS_NSEC        = 53
)
