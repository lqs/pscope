package jvmhsperf

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func connect(pid int) (*net.UnixConn, error) {
	attachFile := filepath.Join(os.TempDir(), fmt.Sprintf(".attach_pid%d", pid))
	f, err := os.OpenFile(attachFile, os.O_CREATE, 0444)
	if err != nil {
		return nil, fmt.Errorf("error creating file %v: %v", attachFile, err)
	}
	_ = f.Close()

	err = syscall.Kill(pid, syscall.SIGQUIT)
	if err != nil {
		return nil, fmt.Errorf("error sending SIGQUIT to %v: %v", pid, err)
	}

	sockFile := filepath.Join(os.TempDir(), fmt.Sprintf(".java_pid%d", pid))
	for i := 0; i < 10; i++ {
		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: sockFile, Net: "unix"})
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		return conn, nil
	}
	return nil, fmt.Errorf("error attaching to process %v", pid)
}

func buildRequest(cmd string, args ...string) []byte {
	var request bytes.Buffer
	request.WriteString("1")
	request.WriteByte(0)
	request.WriteString(cmd)
	request.WriteByte(0)
	for i := 0; i < 3; i++ {
		if i < len(args) {
			request.WriteString(args[i])
		}
		request.WriteByte(0)
	}
	return request.Bytes()
}

func Execute(pid int, cmd string, args ...string) ([]byte, error) {
	conn, err := connect(pid)
	if err != nil {
		return nil, err
	}
	request := buildRequest(cmd, args...)
	if _, err := conn.Write(request); err != nil {
		return nil, fmt.Errorf("error writing to socket: %v", err)
	}

	response, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("error reading from socket: %v", err)
	}

	return response, nil
}
