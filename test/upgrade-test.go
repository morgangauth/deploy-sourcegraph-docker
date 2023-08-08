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
	fmt.Println(cr)
	if err != nil {
	  return err
	}
  
	if !strings.Contains(cr, "deploy-sourcegraph-docker") {
	  return fmt.Errorf("Must run from deploy-sourcegraph-docker repository") 
	}
  
	return nil
  }