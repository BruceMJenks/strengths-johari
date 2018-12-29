package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var dbURL = ""

func TestJohari(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Johari Suite")
}
