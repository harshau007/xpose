package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func DeployDocker(image, port, projectName, projectDir string) error {
	fmt.Printf("Deploying Docker image: %s on port %s for project %s\n", image, port, projectName)
	cmd := exec.Command("docker", "run", "-d", "--name", projectName, "-p", port+":"+port, image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	containerID := strings.TrimSpace(string(output))
	err = Tunnel(port)
	time.Sleep(5 * time.Second)
	SaveProjectInfo(projectName, projectDir, containerID, port, "docker", image)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deployment successful. Container ID: %s\n", containerID)
	return nil
}

func SaveProjectInfo(name, dir, containerID, port, deployType, source string) {
	url, err := ExtractBoreURL("stdoutfile")
	if err != nil {
		fmt.Println(err)
	}
	info := fmt.Sprintf("Name: %s\nContainer ID: %s\nPort: %s\nType: %s\nSource: %s\nPublic URL: %s\n",
		name, containerID, port, deployType, source, url)
	err = os.WriteFile(filepath.Join(dir, "project_info.txt"), []byte(info), 0644)
	if err != nil {
		fmt.Printf("Error saving project info: %v\n", err)
	}
}

func ExtractBoreURL(filename string) (string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a regular expression to match the URL pattern
	re := regexp.MustCompile(`https://[a-f0-9]+\.bore\.digital`)

	// Scan the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Check if the line contains a matching URL
		if match := re.FindString(line); match != "" {
			return match, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// If no URL was found
	return "", fmt.Errorf("no matching URL found in the file")
}

type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func DeployGitHub(repo, port, projectDir string) error {
	projectName, err := getRepoName(repo)
	if err != nil {
		return err
	}

	fmt.Printf("Deploying GitHub repository: %s on port %s for project %s\n", repo, port, projectName)

	// 1. Clone repo to a directory with reponame
	repoPath := filepath.Join(projectDir, projectName)
	dir := filepath.Dir(repoPath)
	fmt.Println("Diretory:", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("git clone %s %s", repo, repoPath))
	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return err
	}

	// 2. Analyze package.json and determine app type and port
	appType, appPort, err := analyzeNodeApp(repoPath)
	if err != nil {
		return err
	}

	// 3. Create a Dockerfile for the specific Node.js application
	dockerfilePath := filepath.Join(repoPath, "Dockerfile")
	if err := createNodeDockerfile(dockerfilePath, appType, appPort); err != nil {
		return err
	}

	// 4. Build the Docker image
	cmd = exec.Command("sh", "-c", fmt.Sprintf("docker build -t %s %s", projectName, repoPath))
	if err := cmd.Run(); err != nil {
		return err
	}

	// 5. Run the Docker container
	cmd = exec.Command("docker", "run", "-d", "--name", projectName, "-p", port+":"+appPort, projectName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	containerID := strings.TrimSpace(string(output))
	err = Tunnel(port)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	SaveProjectInfo(projectName, projectDir, containerID, port, "github", repo)
	fmt.Printf("Deployment successful. Container ID: %s\n", containerID)
	return nil
}

func analyzeNodeApp(repoPath string) (appType string, port string, err error) {
	packageJSONPath := filepath.Join(repoPath, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return "", "", err
	}

	var packageJSON PackageJSON
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return "", "", err
	}

	// Check for Vite
	if _, exists := packageJSON.Scripts["dev"]; exists && strings.Contains(packageJSON.Scripts["dev"], "vite") {
		return "vite", "5173", nil // Vite uses 5173 by default
	}

	// Check for Next.js
	if _, exists := packageJSON.Scripts["dev"]; exists && strings.Contains(packageJSON.Scripts["dev"], "next") {
		return "next", "3000", nil // Next.js uses 3000 by default
	}

	// Default to generic Node.js app
	return "node", "3000", nil
}

func getRepoName(repoURL string) (string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}

	host := u.Hostname()
	if !strings.HasSuffix(host, "github.com") && !strings.HasSuffix(host, "gitlab.com") {
		return "", fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	if !strings.HasPrefix(u.Path, "/") || !strings.Contains(u.Path, "/") {
		return "", fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	pathParts := strings.Split(u.Path, "/")
	repoName := strings.TrimSuffix(pathParts[len(pathParts)-1], ".git")

	return repoName, nil
}

func createNodeDockerfile(path, appType, port string) error {
	var dockerfileTemplate string

	switch appType {
	case "vite":
		dockerfileTemplate = `FROM node:alpine

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

EXPOSE {{.Port}}

CMD ["npm", "run", "preview", "--", "--host", "0.0.0.0", "--port", "{{.Port}}"]
`
	case "next":
		dockerfileTemplate = `FROM node:alpine

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

EXPOSE {{.Port}}

CMD ["npx", "next", "start"]
`
	default:
		dockerfileTemplate = `FROM node:alpine

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

EXPOSE {{.Port}}

CMD ["npm", "start"]
`
	}

	tmpl, err := template.New("dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, struct{ Port string }{Port: port})
	if err != nil {
		return err
	}

	return nil
}
