package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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

// Iterates through an array of version strings and performs the standard upgrade process, checking for errors in the versions and migrations logs tables.
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
	cmd.Dir = "/Users/warrengifford/deploy-sourcegraph-docker/docker-compose"
    if err := streamCommandOutput(cmd); err != nil {
		return fmt.Errorf("failed to run docker-compose up at version %s: %s", version, err)
	}

	err := validateUpgrade()
	if err != nil {
		return fmt.Errorf("failed to validate upgrade at version %s: %s", version, err)
	}

    // Docker compose down
    cmd = exec.Command("docker-compose", "down", "--remove-orphans")
    cmd.Dir = "/Users/warrengifford/deploy-sourcegraph-docker/docker-compose"
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

func validateUpgrade (version string) error {
	// Validate the database version was set correctly.
	cmd := exec.Command("docker", "exec", "-it", "pgsql", "psql", "-U", "sg", "-c", "'SELECT version FROM versions;'")
	dbv, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to validate upgraded version %s: %s", version, err)
	}
	dbvStr := string(dbv)	
	if dbvStr != version {
		return fmt.Errorf("Database version %s not set during upgrade. Table versions.version = %s", version, dbvStr)
	}
	// Check for failed migrations in migration_logs table.
	cmd = exec.Command("docker", "exec", "-it", "pgsql", "psql", "-U", "sg", "-c", "'SELECT COUNT(*) FROM migration_logs WHERE success = false;'")
    fmc, err := cmd.CombinedOutput()
	failedMigrations, err := strconv.Atoi(string(fmc))
	if err != nil {
		return fmt.Errorf("Error parsing failed migrations count: %s", err)
	}
	if failedMigrations > 0 {
		return fmt.Errorf("Failed migrations found after upgrade to version %s, database marked dirty.", version)
	}
	return nil
}