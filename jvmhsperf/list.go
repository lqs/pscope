package jvmhsperf

import (
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lqs/pscope/common"
	"github.com/shirou/gopsutil/v3/process"
)

func ListProcesses() []common.Process {
	var pids []int
	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "hsperfdata_") || !entry.IsDir() {
			continue
		}
		subEntries, err := os.ReadDir(filepath.Join(os.TempDir(), entry.Name()))
		if err != nil {
			continue
		}
		for _, subEntry := range subEntries {
			pid, err := strconv.Atoi(subEntry.Name())
			if err != nil {
				continue
			}
			pids = append(pids, pid)
		}
	}

	var processes []common.Process
	for _, pid := range pids {
		proc, err := process.NewProcess(int32(pid))
		if err != nil {
			continue
		}
		argv, _ := proc.CmdlineSlice()
		var cmdline string
		if len(argv) > 2 && path.Base(argv[0]) == "java" {
			// TODO: parse java cmdline and extract main class name
			cmdline = argv[len(argv)-1] // hope that the last arg is the main class
		} else {
			cmdline, _ = proc.Cmdline()
		}
		createTime, _ := proc.CreateTime()
		processes = append(processes, common.Process{
			Type:         common.ProcessTypeJava,
			PID:          pid,
			BuildVersion: "java",
			Path:         cmdline,
			Agent:        true,
			StartTime:    time.UnixMilli(createTime),
		})
	}

	return processes
}
