# GoShell — Custom POSIX-style Shell (Go)

A lightweight, interactive Unix-like shell built from scratch in Go.
This project demonstrates `process execution`, `command parsing`, `I/O redirection`, `pipelines`, and `interactive terminal behavior` — core concepts behind real operating system shells.


# Run with Docker (Recommended)

**Pull and run instantly:**

```sh
docker pull niraj1910/codecrafters-shell:latest
docker run -it niraj1910/codecrafters-shell:latest
```

**Or build locally:**

```sh
git clone <your-repo-url>
cd <repo>

docker build -t goshell .
docker run -it goshell
```

# Local Run

Requirements: **Go 1.25+**

```sh
go run app/main.go
```

# Features

**Interactive REPL**

- Persistent shell prompt ($)

- Command editing via readline

- Handles empty input and read errors gracefully

- Built-in tab auto-completion for shell commands


# Built-in Commands

```sh
Command	       Description
cd <dir>	      #Change working directory
pwd	         #Print current working directory
echo	         #Print arguments
type <cmd>	   #Check if command is builtin or external
history	      #Display command history
exit	         #Exit shell (persists history)
```

# Command History (File-backed)

**Supports advanced history operations:**

```sh

history                # show all history
history 10             # last 10 commands
history -r file.txt    # load history
history -w file.txt    # overwrite history file
history -a file.txt    # append new commands

```

**Persistent history using environment variable:**

```sh
HISTFILE=history.txt ./shell
```


# I/O Redirection

**Supported:**

```sh
echo hello > out.txt
ls > output.txt
command 2> error.txt
```

Separate stdout and stderr redirection

Safe file handling per command execution


# Pipeline Support

**Execute chained commands:**

```sh 
cat file.txt | grep foo | wc -l
```

Commands are parsed and executed with pipe chaining.



# Architecture Overview

**Core components:**

- REPL loop with readline

- Command tokenizer & parser

- Pipeline splitter

- Built-in command dispatcher

- External process runner

- History manager

- Output redirection handler

- Execution flow:

```sh
User Input
   ↓
Parse → Detect pipeline?
   ↓
Built-in OR External
   ↓
Process execution / piping / redirection
```

# Example Session

```sh
$ pwd
/home/app

$ type exit
exit is a shell builtin

$ echo hello > file.txt

$ history
1 pwd
2 type exit
3 echo hello > file.txt

```

# Tech Stack

1. Go

2. readline

3. Docker (multi-stage build)

4. Linux process model


# Why This Project Matters

*This implementation demonstrates:*

- Systems programming fundamentals

- Process management in Linux

- File descriptor and stream control

- Command parsing and REPL design

- OS-level execution behavior

- Production containerization

- Go for system tooling

*Relevant for roles in:*

- Backend Engineering

- Platform / Infrastructure Engineering

- DevTools / CLI Development

- Systems Programming (Go)
