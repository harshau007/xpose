package services

import (
	"bufio"

	// "bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/jkuri/bore/client"
)

func Tunnel(port string) error {
	fmt.Println("Serving port", port)
	boreCmd := fmt.Sprintf("bore -s bore.digital -p 2200 -ls localhost -lp %s > stdoutfile 2> stderrfile &", port)
	cmd := exec.Command("sh", "-c", boreCmd)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	// buf := new(bytes.Buffer)
	// _, _ = io.Copy(buf, stdout)
	// fmt.Println(buf.String())
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}

	// Create a channel to signal when we're done reading logs
	done := make(chan bool)

	// Function to read from a pipe and print logs
	logReader := func(reader io.Reader, prefix string) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Printf("%s: %s\n", prefix, scanner.Text())
		}
	}

	// Start goroutines to read stdout and stderr
	go func() {
		logReader(stdout, "STDOUT")
		done <- true
	}()
	go func() {
		logReader(stderr, "STDERR")
		done <- true
	}()

	// Wait for both goroutines to finish
	<-done
	<-done

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command finished with error: %v", err)
	}

	return nil
}

func ReadOutput() {
	url, err := ExtractBoreURL("stdoutfile")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(url)
}

type BoreClient struct {
	RemoteServer string
	RemotePort   int
	LocalServer  string
	LocalPort    int
	BindPort     int
	ID           string
	KeepAlive    bool
}

func UsingBore(autoReconnect bool, config BoreClient) error {
	client := client.NewBoreClient(client.Config{
		RemoteServer: config.RemoteServer,
		RemotePort:   config.RemotePort,
		LocalServer:  config.LocalServer,
		LocalPort:    config.LocalPort,
		BindPort:     config.BindPort,
		ID:           config.ID,
		KeepAlive:    config.KeepAlive,
	})

connect:
	if err := client.Run(); err != nil {
		if !autoReconnect {
			return err
		}
		log.Println("connection failed due: ", err.Error(), "reconnecting in 5s...")
		time.Sleep(time.Second * 5)
		goto connect
	}
	return nil
}

func ReadLine(r io.Reader, lineNum int) (line string, lastLine int, err error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lastLine++
		if lastLine == lineNum {
			// you can return sc.Bytes() if you need output in []bytes
			return sc.Text(), lastLine, sc.Err()
		}
	}
	return line, lastLine, io.EOF
}
