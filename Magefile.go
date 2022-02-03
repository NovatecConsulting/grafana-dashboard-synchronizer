//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	// mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

func getExecutableName(os string, arch string) string {
	exeName := fmt.Sprintf("%s_%s_%s", "grafana-dashboard-sync", os, arch)
	if os == "windows" {
		exeName = fmt.Sprintf("%s.exe", exeName)
	}
	return exeName
}

// A build step that requires additional params, or platform specific steps for example
func buildPlatform(os string, arch string) error {
	exeName := getExecutableName(os, arch)

	envMap := make(map[string]string)

	envMap["GOARCH"] = arch
	envMap["GOOS"] = os

	// TODO: Change to sh.RunWithV once available.
	return sh.RunWith(envMap, "go", "build", "-o", filepath.Join("dist", exeName), "./pkg")
}

func BuildWindows() error {
	return buildPlatform("windows", "amd64")
}

func BuildLinux() error {
	return buildPlatform("linux", "amd64")
}

func BuildLinuxARM() error {
	return buildPlatform("linux", "arm")
}

func BuildLinuxARM64() error {
	return buildPlatform("linux", "arm64")
}

func BuildDarwin() error {
	return buildPlatform("darwin", "amd64")
}

func BuildDarwinARM64() error {
	return buildPlatform("darwin", "arm64")
}

func BuildAll() { //revive:disable-line
	mg.Deps(Clean)

	fmt.Println("Building all platforms...")

	mg.Deps(BuildWindows, BuildLinux, BuildLinuxARM, BuildLinuxARM64, BuildDarwin, BuildDarwinARM64)
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("dist")
}

var Default = BuildAll
