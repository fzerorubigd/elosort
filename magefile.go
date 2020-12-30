// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Build build the elobot binary.
func Build() error {
	gocmd := mg.GoCmd()
	return sh.RunV(gocmd, "build", "-o", "elobot", "./cmd/elobot")
}

// Docker build the docker file.
func Docker() error {
	return sh.RunV("docker", "build", ".", "--file", "Dockerfile", "--tag", "elobot")
}
