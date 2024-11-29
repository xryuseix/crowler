package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"

	"sync"
	"time"
)

type dirInfo struct {
	name string
	done bool
}

func getDirectories(wg *sync.WaitGroup, dirChan chan dirInfo, sigChan chan os.Signal) {
	defer wg.Done()
	cmdString := fmt.Sprintf("ssh -i %s %s@%s ls -l %s | grep '^d' | awk '{print $9}' | tr '\n' ' '", env.IdentityPath, env.ServerUser, env.ServerIP, env.RemotePath)
	cmd := exec.Command("/bin/sh", "-c", cmdString)

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("[ERROR] Failed to list directories: %v\n", err)
	}

	dirs := strings.Split(strings.Trim(string(output), " "), " ")
	log.Printf("[INFO] Found %d directories\n", len(dirs))

L:
	for _, dir := range dirs {
		select {
		case <-sigChan:
			log.Println("[INFO] Ctrl+C received, stopping...")
			break L
		default:
			dirChan <- dirInfo{name: dir, done: false}
		}
	}
	close(dirChan)
}

// func downloadDirectory(name string) {
//     tarCmdStr := fmt.Sprintf("tar czf %s.tar.gz -C %s %s", name, outDir, name)
//     tarCmd := exec.Command("/bin/sh", "-c", tarCmdStr)
//     err := tarCmd.Run()
//     if err != nil {
//         log.Printf("Failed to create tarball for directory %s: %v\n", name, err)
//         return
//     }

//     scpCmdStr := fmt.Sprintf("scp -i %s %s@%s:%s.tar.gz .", identityFile, serverUser, serverIP, name)
//     scpCmd := exec.Command("/bin/sh", "-c", scpCmdStr)
//     err = scpCmd.Run()
//     if err != nil {
//         log.Printf("Failed to download tarball for directory %s: %v\n", name, err)
//         return
//     }

//     fmt.Printf("Downloaded directory: %s\n", name)
// }

func downloadDirectory(name string, current int) {
	log.Printf("[%d] Downloading directory(1/3): %s\n", current, name)
	time.Sleep(1 * time.Second)
	log.Printf("[%d] Downloading directory(2/3): %s\n", current, name)
	time.Sleep(1 * time.Second)
	log.Printf("[%d] Downloading directory(3/3): %s\n", current, name)
	time.Sleep(1 * time.Second)
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
