package integration

import (
	"os"

	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// CoreDNSSuite.
type CoreDNSSuite struct{ BaseSuite }

func (s *CoreDNSSuite) SetUpSuite(c *check.C) {
	requiredImages := []string{
		"containous/whoami:v1.0.1",
		"coredns/coredns:1.2.6",
		"coredns/coredns:1.3.1",
		"coredns/coredns:1.4.0",
		"coredns/coredns:1.5.2",
		"coredns/coredns:1.6.3",
		"coredns/coredns:1.7.0",
		"giantswarm/tiny-tools:3.9",
	}

	s.startk3s(c, requiredImages)
	s.startWhoami(c)
	s.installTinyToolsMaesh(c)
	s.createResources(c, "testdata/smi/crds/")
}

func (s *CoreDNSSuite) TearDownSuite(c *check.C) {
	s.stopK3s()
}

func (s *CoreDNSSuite) TestCoreDNSVersionSafe(c *check.C) {
	testCases := []struct {
		desc          string
		version       string
		expectedError bool
	}{
		{
			desc:          "CoreDNS 1.2.6",
			version:       "1.2.6",
			expectedError: true,
		},
		{
			desc:          "CoreDNS 1.3.1",
			version:       "1.3.1",
			expectedError: false,
		},
		{
			desc:          "CoreDNS 1.4.0",
			version:       "1.4.0",
			expectedError: false,
		},
	}

	for _, test := range testCases {
		s.createResources(c, "testdata/coredns/corednssafe.yaml")
		s.WaitForCoreDNS(c)

		c.Logf("Testing compatibility with %s", test.desc)
		s.setCoreDNSVersion(c, test.version)

		cmd := s.maeshPrepareWithArgs()
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()

		c.Log(string(output))

		if test.expectedError {
			c.Assert(err, checker.NotNil)
		} else {
			c.Assert(err, checker.IsNil)
		}

		s.deleteResources(c, "testdata/coredns/corednssafe.yaml")
	}
}

func (s *CoreDNSSuite) TestCoreDNSVersion(c *check.C) {
	testCases := []struct {
		desc    string
		version string
	}{
		{
			desc:    "CoreDNS 1.5.2",
			version: "1.5.2",
		},
		{
			desc:    "CoreDNS 1.6.3",
			version: "1.6.3",
		},
		{
			desc:    "CoreDNS 1.7.0",
			version: "1.7.0",
		},
	}

	for _, test := range testCases {
		s.createResources(c, "testdata/coredns/coredns.yaml")
		s.WaitForCoreDNS(c)

		c.Logf("Testing compatibility with %s", test.desc)
		s.setCoreDNSVersion(c, test.version)

		cmd := s.maeshPrepareWithArgs()
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()

		c.Log(string(output))
		c.Assert(err, checker.IsNil)

		s.deleteResources(c, "testdata/coredns/coredns.yaml")
	}
}

func (s *CoreDNSSuite) TestCoreDNSDig(c *check.C) {
	testCases := []struct {
		desc    string
		version string
	}{
		{
			desc:    "CoreDNS 1.6.3",
			version: "1.6.3",
		},
		{
			desc:    "CoreDNS 1.7.0",
			version: "1.7.0",
		},
	}

	for _, test := range testCases {
		s.createResources(c, "testdata/coredns/coredns.yaml")
		s.WaitForCoreDNS(c)

		c.Logf("Testing dig with %s", test.desc)
		s.setCoreDNSVersion(c, test.version)

		cmd := s.startMaeshBinaryCmd(c, false, false)

		err := cmd.Start()
		c.Assert(err, checker.IsNil)

		pod := s.getToolsPodMaesh(c)
		c.Assert(pod, checker.NotNil)

		s.digHost(c, pod.Name, pod.Namespace, "whoami.whoami.maesh")
		s.stopMaeshBinary(c, cmd.Process)

		s.deleteResources(c, "testdata/coredns/coredns.yaml")
	}
}
