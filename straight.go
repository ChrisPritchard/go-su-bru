package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 3 {
		log.Fatal("args: username password")
	}

	username := os.Args[1]
	password := os.Args[2]

	pty, tty, err := open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()
	defer pty.Close()

	c := exec.Command("su", "-c", "id", username)
	defer c.Wait()
	c.Stdout = tty
	c.Stderr = tty
	c.Stdin = tty

	if err := c.Start(); err != nil {
		log.Fatal(err)
	}

	buffer := make([]byte, 100)
	n, _ := pty.Read(buffer) // text 'Password: '
	log.Print(string(buffer[:n]))
	pty.Write([]byte(password + "\n"))
	log.Println(password)
	n, _ = pty.Read(buffer) // new line
	log.Println(string(buffer[:n]))
	n, _ = pty.Read(buffer) // either id text or failure
	log.Println(string(buffer[:n]))

	if n > 0 && strings.HasPrefix(string(buffer), "uid=") {
		log.Printf("success with %s\n", password)
		os.Exit(0)
	}

	c = exec.Command("su", "-c", "id", username)
	defer c.Wait()
	c.Stdout = tty
	c.Stderr = tty
	c.Stdin = tty

	if err := c.Start(); err != nil {
		log.Fatal(err)
	}

	buffer = make([]byte, 100)
	n, _ = pty.Read(buffer) // text 'Password: '
	log.Print(string(buffer[:n]))
	pty.Write([]byte(password + "\n"))
	log.Println(password)
	n, _ = pty.Read(buffer) // new line
	log.Println(string(buffer[:n]))
	n, _ = pty.Read(buffer) // either id text or failure
	log.Println(string(buffer[:n]))

	if n > 0 && strings.HasPrefix(string(buffer), "uid=") {
		log.Printf("success with %s\n", password)
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
