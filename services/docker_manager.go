package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type DockerManager struct {
	containerName string
	imageName     string
	port          string
	mu            sync.Mutex
	isRunning     bool
	lastUsed      time.Time
	stopTimer     *time.Timer
}

var dockerManager *DockerManager
var dockerOnce sync.Once

func GetDockerManager() *DockerManager {
	dockerOnce.Do(func() {
		dockerManager = &DockerManager{
			containerName: "wardrobe-local-ai",
			imageName:     "wardrobe-local-ai:latest",
			port:          "8081",
			isRunning:     false,
		}
	})
	return dockerManager
}

func (d *DockerManager) StartContainer() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isRunning {
		log.Printf("DockerManager: AI container already running, resetting timer")
		d.resetStopTimerLocked()
		return nil
	}

	log.Printf("DockerManager: Starting AI container...")

	cmd := exec.Command("docker", "run", "-d",
		"--name", d.containerName,
		"-p", d.port+":8081",
		"-e", "PYTHONUNBUFFERED=1",
		d.imageName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("DockerManager: Failed to start container: %v, output: %s", err, string(output))
		return fmt.Errorf("failed to start container: %v", err)
	}

	if err := d.waitForReadyLocked(60 * time.Second); err != nil {
		d.stopContainerLocked()
		return fmt.Errorf("container not ready: %v", err)
	}

	d.isRunning = true
	d.lastUsed = time.Now()
	d.resetStopTimerLocked()

	log.Printf("DockerManager: AI container started successfully")
	return nil
}

func (d *DockerManager) waitForReadyLocked(timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", d.port))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout waiting for container")
}

func (d *DockerManager) resetStopTimerLocked() {
	if d.stopTimer != nil {
		d.stopTimer.Stop()
	}
	d.stopTimer = time.AfterFunc(5*time.Minute, func() {
		d.StopContainer()
	})
}

func (d *DockerManager) StopContainer() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isRunning {
		return nil
	}

	log.Printf("DockerManager: Stopping AI container (idle timeout)...")

	if d.stopTimer != nil {
		d.stopTimer.Stop()
	}

	err := d.stopContainerLocked()
	if err == nil {
		d.isRunning = false
		log.Printf("DockerManager: AI container stopped successfully")
	}

	return err
}

func (d *DockerManager) stopContainerLocked() error {
	cmd := exec.Command("docker", "stop", d.containerName)
	cmd.Run()

	cmd = exec.Command("docker", "rm", d.containerName)
	cmd.Run()

	return nil
}

func (d *DockerManager) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.isRunning
}

func (d *DockerManager) CallRecognize(imageBase64 string) (map[string]interface{}, error) {
	if err := d.StartContainer(); err != nil {
		return nil, fmt.Errorf("failed to start AI container: %v", err)
	}

	url := fmt.Sprintf("http://localhost:%s/recognize", d.port)

	reqBody := map[string]interface{}{
		"image": imageBase64,
	}
	jsonBody, _ := json.Marshal(reqBody)

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	d.mu.Lock()
	d.lastUsed = time.Now()
	d.resetStopTimerLocked()
	d.mu.Unlock()

	return result, nil
}
