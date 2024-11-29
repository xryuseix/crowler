package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type dirInfo struct {
	name string
	done bool
}

func execRemoteCommand(_cmd string) string {
	sshCmd := fmt.Sprintf("ssh -i %s %s@%s %s", env.IdentityPath, env.ServerUser, env.ServerIP, _cmd)
	cmd := exec.Command("/bin/sh", "-c", sshCmd)

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("[ERROR] Failed to list directories: %v\n", err)
	}
	return string(output)
}

func getDirectories(wg *sync.WaitGroup, dirChan chan dirInfo, sigChan chan os.Signal) {
	defer wg.Done()
	cmd := fmt.Sprintf("ls -l %s | grep '^d' | awk '{print $9}' | tr '\n' ' '", env.RemotePath)
	output := execRemoteCommand(cmd)

	dirs := strings.Split(strings.Trim(string(output), " "), " ")
	log.Printf("[INFO] Found %d directories\n", len(dirs))

L:
	for _, dir := range dirs {
		select {
		case <-sigChan:
			log.Println("[INFO] Ctrl+C received, will be stopping...")
			break L
		default:
			dirChan <- dirInfo{name: dir, done: false}
		}
	}
	close(dirChan)
}

func downloadDirectory(name string, current int) {
	fmt.Printf("[%d] Downloading directory: %s\n", current, name)
	rp := filepath.Join(env.LocalPath, name)
	scpCmdStr := fmt.Sprintf("'tar zcf - %s && rm -rf %s' | tar zxf -", rp, rp)
	execRemoteCommand(scpCmdStr)

	time.Sleep(3 * time.Second)
}

func init() {
	LoadEnv()
	if err := exec.Command("mkdir", "-p", env.LocalPath).Run(); err != nil {
		log.Fatalf("[ERROR] Failed to create the local directory: %v\n", err)
	}
}

type Progress struct {
	current int
	mu      sync.Mutex
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	progress := Progress{current: 0}

	wg := &sync.WaitGroup{}
	dirChan := make(chan dirInfo)
	wg.Add(1)

	go getDirectories(wg, dirChan, sigChan)

	sem := make(chan bool, min(env.MaxWorkers, runtime.NumCPU())) // semaphore
	for info := range dirChan {
		if !info.done {
			sem <- true
			progress.mu.Lock()
			progress.current++
			go func(dirName string, current int) {
				downloadDirectory(dirName, current)
				<-sem
			}(info.name, progress.current)
			progress.mu.Unlock()
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	wg.Wait()
}
