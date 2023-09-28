package gops

// This file is copied and modified from "gops" project.

import (
	"context"
	"fmt"
	"io"
	"net"
)

func pidToAddr(pid int) (*net.TCPAddr, error) {
	port, err := GetPort(pid)
	if err != nil {
		return nil, fmt.Errorf("couldn't get port for PID %v: %v", pid, err)
	}
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+port)
	return addr, nil
}

func Cmd(ctx context.Context, pid int, c byte, params ...byte) ([]byte, error) {
	conn, err := cmdLazy(pid, c, params...)
	if err != nil {
		return nil, fmt.Errorf("couldn't get port by PID: %v", err)
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	all, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	return all, nil
}

func cmdLazy(pid int, c byte, params ...byte) (io.ReadCloser, error) {
	addr, err := pidToAddr(pid)

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	buf := []byte{c}
	buf = append(buf, params...)
	if _, err := conn.Write(buf); err != nil {
		return nil, err
	}
	return conn, nil
}
