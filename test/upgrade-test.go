package main

import (
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
    } else if (testType == "multiversion" || testType == "mvu") {
        testMultiversionUpgrade(versions)
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
	versionRegex, nil := regexp.Compile(`^v?(\d+\.\d+\.\d+)$`)
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
	fmt.Println(crs)
  
	if !strings.Contains(crs, "deploy-sourcegraph-docker") {
	  return fmt.Errorf("Must run from deploy-sourcegraph-docker repository") 
	}
  
	return nil
  }

  func performStandardUpgrade(versions []string) error {

  // Prune docker volumes
  if err := exec.Command("docker", "volume", "prune", "-f").Run(); err != nil {
    return fmt.Errorf("error pruning docker volumes: %s", err)
  }

  for i, version := range versions {

    // Checkout version tag
    if !strings.HasPrefix(version, "v") {
		versions[i] = "v" + version
	}
    if err := exec.Command("git", "checkout", version).Run(); err != nil {
      return fmt.Errorf("error checking out %s: %s", version, err)
    }

    // Docker compose up
    cmd := exec.Command("docker-compose", "up", "-d")
    cmd.Dir = "deploy-sourcegraph-docker/docker-compose" 
    if err := cmd.Run(); err != nil {
      return fmt.Errorf("error running docker-compose up for %s: %s", version, err)
    }

    if i < len(versions)-1 {
      // Docker compose down
      cmd := exec.Command("docker-compose", "down", "--remove-orphans")
      cmd.Dir = "deploy-sourcegraph-docker/docker-compose"
      if err := cmd.Run(); err != nil {
        return fmt.Errorf("error running docker-compose down for %s: %s", version, err)  
      }
    }

  }

  return nil

}
  