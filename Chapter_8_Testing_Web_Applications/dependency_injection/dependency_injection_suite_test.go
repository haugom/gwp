package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDependencyInjection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DependencyInjection Suite")
}
