package sshterminal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTerminal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SSH Terminal Suite")
}
