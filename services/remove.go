package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func Remove(ids []string, projectDir string) error {
	transaction, err := NewTransaction(FilePath(projectDir))
	if err != nil {
		return err
	}
	defer transaction.rollback()

	for _, id := range ids {
		if entry, found := transaction.ReadEntryById(id); found {
			if strings.EqualFold(entry.Type, "github") {
				projPath := filepath.Join(projectDir, entry.Name)
				err := os.RemoveAll(projPath)
				if err != nil {
					log.Fatal("Error Deleting Project folder", err)
				}
			}
			pid, err := strconv.Atoi(entry.PID)
			if err != nil {
				log.Fatalf("No process with %s Found. Err %s", entry.PID, err)
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				log.Fatal(err)
			}
			process.Kill()

			contKill := fmt.Sprintf("docker kill %s && docker rm %s", entry.ContainerID, entry.ContainerID)
			cmd := exec.Command("sh", "-c", contKill)
			_, err = cmd.CombinedOutput()
			if err != nil {
				return err
			}

			transaction.DeleteEntry(id)
		}
	}

	if err := transaction.commit(); err != nil {
		return err
	}
	return nil
}
