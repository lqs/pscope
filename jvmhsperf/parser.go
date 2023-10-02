package jvmhsperf

import (
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/lqs/pscope/common"
)

// "http-nio-8080-exec-1" #20 daemon prio=5 os_prio=0 cpu=126.11ms elapsed=144.34s tid=0x00007f9048f3c290 nid=0xc038 waiting on condition  [0x00007f8faf7f9000]
var headerRegex = regexp.MustCompile(`^"(.+)" #(\d+) (?:.+ )?cpu=(\d+(?:\.\d+)?m?s) (?:.+ )?nid=(0x[0-9a-f]+) (.+?) +(\[(0x[0-9a-f]+)])?$`)

// java.lang.Thread.State: TIMED_WAITING (sleeping)
var threadStateRegex = regexp.MustCompile(`^java\.lang\.Thread\.State: (.+)$`)

// at some.package.Class$SubClass.method(RequestMappingHandlerAdapter.java:878)
var frameRegex = regexp.MustCompile(`^at (?:([a-zA-Z0-9.]*?)\.)?([a-zA-Z0-9$]+)\.([a-zA-Z0-9]+)\((.+):(\d+)\)$`)

func ParseJavaThreadDump(threadDump []byte) []*common.CallStack {
	var threads []*common.CallStack
	reader := bufio.NewReader(bytes.NewReader(threadDump))
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			threadId, _ := strconv.Atoi(matches[2])
			thread := &common.CallStack{
				Id:    threadId,
				Name:  matches[1],
				State: matches[5],
			}
			threads = append(threads, thread)
		} else if matches := threadStateRegex.FindStringSubmatch(line); matches != nil {
			thread := threads[len(threads)-1]
			thread.State = matches[1]
		} else if matches := frameRegex.FindStringSubmatch(line); matches != nil {
			frame := &common.Frame{
				Package: matches[1],
				Func:    matches[2] + "." + matches[3],
				File:    matches[4],
			}
			thread := threads[len(threads)-1]
			thread.Frames = append(thread.Frames, frame)
		} else if strings.Contains(line, "- Coroutine dump -") {
			// TODO: add coroutine support
			break
		}
	}
	return threads
}
