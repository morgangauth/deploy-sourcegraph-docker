package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
    testType := os.Args[1]
	versions := os.Args[2:]

	if err := checkRepo(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	  }

    if testType == "standard" {
        testStandardUpgrade(versions)
		if err := performStandardUpgrade(versions); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
    } else if (testType == "multiversion" || testType == "mvu") {
        testMultiversionUpgrade(versions)
    } else {
		fmt.Println("Error: Must declare testType as 'standard' or 'multiversion'")
	}
}

func testStandardUpgrade(versions []string) error {
	if err := checkVersion(versions); err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Println("Standard Upgrade Test")
	for _, version := range versions {
		fmt.Println(version)
	}
	return nil
}

func testMultiversionUpgrade(versions []string) error {
	if err := checkVersion(versions); err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Println("Multi-version Upgrade Test")
	for _, version := range versions {
		fmt.Println(version)
	}
	return nil
}

func checkVersion(versions []string) error {
	versionRegex, nil := regexp.Compile(`^v?(\d+\.\d+\.\d+)`)
	for _, version := range versions {
		if !versionRegex.MatchString(version) {
			return fmt.Errorf("invalid version %s", version)
		}
	}
	return nil
}

func checkRepo() error {
	cr, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return fmt.Errorf("git rev-parse --show-toplevel failed with:  %s", err)///
	  }
	
	crs := strings.TrimSpace(string(cr))
  
	if !strings.Contains(crs, "deploy-sourcegraph-docker") {
	  return fmt.Errorf("Must run from deploy-sourcegraph-docker repository") 
	}
  
	return nil
}

func performStandardUpgrade(versions []string) error {
  // Prune docker volumes
  cmd := exec.Command("docker", "volume", "prune", "-f")
  fmt.Println("Pruning docker volumes...")
  if err := streamCommandOutput(cmd); err != nil {
	return fmt.Errorf("failed to prune docker volumes: %s", err)
  }

  for i, version := range versions {

    // Checkout version tag
    if !strings.HasPrefix(version, "v") {
		versions[i] = "v" + version
	}
    cmd := exec.Command("git", "checkout", version)
	if err := streamCommandOutput(cmd); err != nil {
		return fmt.Errorf("failed to checkout version %s: %s", version, err)
	}

    // Docker compose up
    cmd = exec.Command("docker-compose", "up", "-d")
	cmd.Dir = "deploy-sourcegraph-docker/docker-compose"
    if err := streamCommandOutput(cmd); err != nil {
		return fmt.Errorf("failed to run docker-compose up at version %s: %s", version, err)
	}

    // Docker compose down
    cmd = exec.Command("docker-compose", "down", "--remove-orphans")
    cmd.Dir = "deploy-sourcegraph-docker/docker-compose"
	if err := streamCommandOutput(cmd); err != nil {
		return fmt.Errorf("failed to run docker-compose down at version %s: %s", version, err)
	}
  }
  return nil
}

func streamCommandOutput(cmd *exec.Cmd) error {
	stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
      return fmt.Errorf("error starting command: %s", err)
    }
	stdScanner := bufio.NewScanner(stdout)
    for stdScanner.Scan() {
      fmt.Println(stdScanner.Text()) 
    }
    errScanner := bufio.NewScanner(stderr)
    for errScanner.Scan() {
      fmt.Println(errScanner.Text()) 
    }
    if err := cmd.Wait(); err != nil {
	  return err
    }
	return nil
}