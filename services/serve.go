package services

import (
	"bufio"
	"bytes"
	"strings"

	// "bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/jkuri/bore/client"
)

func Serve(port string) (string, error) {
	fmt.Println("Serving port", port)
	boreCmd := fmt.Sprintf("bore -s bore.digital -p 2200 -ls localhost -lp %s > stdoutfile 2> stderrfile & echo $!", port)
	cmd := exec.Command("sh", "-c", boreCmd)

	// Capture the command's output
	var out bytes.Buffer
	cmd.Stdout = &out

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting command: %v", err)
	}

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("command finished with error: %v", err)
	}

	pid := out.String()

	return strings.TrimSpace(pid), nil
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
