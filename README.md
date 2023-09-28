# pscope

## Introduction
**pscope** is an interactive tool designed to examine currently running Go and Java processes. It can be considered as a more user-friendly variant of `gops` and `jps`, with a text-based user interface.

## Features
* List all running Go and Java processes
* Show detailed information and runtime state of a process
* Show stack trace of a goroutine or thread
* Show heap profile of a process
* ... and more to come!

## Installation

### From source

```sh
$ go install github.com/lqs/pscope@latest
```

### From pre-built binaries
(TODO)

### Embed in your Docker image
You can embed `pscope` in your Docker image to easily use it in your container, even within a Kubernetes pod.

(TODO)