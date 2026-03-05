package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}

var _ = Describe("ParseWatchNamespaces", func() {
	DescribeTable("Parses correctly",
		func(input string, wantKeys []string, wantNil bool) {
			got := parseWatchNamespaces(input)
			if wantNil {
				Expect(got).To(BeNil())
			} else {
				Expect(got).NotTo(BeNil())
				Expect(got).To(HaveLen(len(wantKeys)))
				for _, k := range wantKeys {
					Expect(got).To(HaveKey(k))
				}
			}
		},
		Entry("empty string returns nil (cluster-wide)", "", nil, true),
		Entry("whitespace-only returns nil (cluster-wide)", "  ", nil, true),
		Entry("comma-only returns nil (cluster-wide)", ",,,", nil, true),
		Entry("mixed blank entries returns nil (cluster-wide)", " , , ", nil, true),
		Entry("single namespace", "ns-a", []string{"ns-a"}, false),
		Entry("single namespace with surrounding spaces", "  ns-a  ", []string{"ns-a"}, false),
		Entry("multiple namespaces", "ns-a,ns-b,ns-c", []string{"ns-a", "ns-b", "ns-c"}, false),
		Entry("multiple namespaces with spaces", " ns-a , ns-b , ns-c ", []string{"ns-a", "ns-b", "ns-c"}, false),
		Entry("duplicates are deduplicated by map", "ns-a,ns-a,ns-b", []string{"ns-a", "ns-b"}, false),
		Entry("blank entries mixed with valid are skipped", "ns-a,,ns-b, ,ns-c", []string{"ns-a", "ns-b", "ns-c"}, false),
	)
})
