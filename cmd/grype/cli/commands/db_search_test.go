package commands

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anchore/grype/cmd/grype/cli/commands/internal/dbsearch"
	"github.com/anchore/grype/cmd/grype/cli/options"
)

func TestDBSearchMatchOptionsApplyArgs(t *testing.T) {
	testCases := []struct {
		name               string
		args               []string
		expectedPackages   []string
		expectedVulnIDs    []string
		expectedErrMessage string
	}{
		{
			name:             "empty arguments",
			args:             []string{},
			expectedPackages: []string{},
			expectedVulnIDs:  []string{},
		},
		{
			name: "valid cpe",
			args: []string{"cpe:2.3:a:vendor:product:1.0:*:*:*:*:*:*:*"},
			expectedPackages: []string{
				"cpe:2.3:a:vendor:product:1.0:*:*:*:*:*:*:*",
			},
			expectedVulnIDs: []string{},
		},
		{
			name: "valid purl",
			args: []string{"pkg:npm/package-name@1.0.0"},
			expectedPackages: []string{
				"pkg:npm/package-name@1.0.0",
			},
			expectedVulnIDs: []string{},
		},
		{
			name:             "valid vulnerability IDs",
			args:             []string{"CVE-2023-0001", "GHSA-1234", "ALAS-2023-1234"},
			expectedPackages: []string{},
			expectedVulnIDs: []string{
				"CVE-2023-0001",
				"GHSA-1234",
				"ALAS-2023-1234",
			},
		},
		{
			name: "mixed package and vulns",
			args: []string{"cpe:2.3:a:vendor:product:1.0:*:*:*:*:*:*:*", "CVE-2023-0001"},
			expectedPackages: []string{
				"cpe:2.3:a:vendor:product:1.0:*:*:*:*:*:*:*",
			},
			expectedVulnIDs: []string{
				"CVE-2023-0001",
			},
		},
		{
			name: "plain package name",
			args: []string{"package-name"},
			expectedPackages: []string{
				"package-name",
			},
			expectedVulnIDs: []string{},
		},
		{
			name: "invalid PostLoad error for Package",
			args: []string{"pkg:npm/package-name@1.0.0", "cpe:invalid"},
			expectedPackages: []string{
				"pkg:npm/package-name@1.0.0",
			},
			expectedErrMessage: "invalid CPE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &dbSearchMatchOptions{
				Vulnerability: options.DBSearchVulnerabilities{},
				Package:       options.DBSearchPackages{},
			}

			err := opts.applyArgs(tc.args)

			if tc.expectedErrMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMessage)
				return
			}

			require.NoError(t, err)
			if d := cmp.Diff(tc.expectedPackages, opts.Package.Packages, cmpopts.EquateEmpty()); d != "" {
				t.Errorf("unexpected package specifiers: %s", d)
			}
			if d := cmp.Diff(tc.expectedVulnIDs, opts.Vulnerability.VulnerabilityIDs, cmpopts.EquateEmpty()); d != "" {
				t.Errorf("unexpected vulnerability specifiers: %s", d)
			}

		})
	}
}
func TestV5Namespace(t *testing.T) {
	// provider input should be derived from the Providers table:
	// +------------+---------+---------------+----------------------------------+------------------------+
	// | id         | version | processor     | date_captured                    | input_digest           |
	// +------------+---------+---------------+----------------------------------+------------------------+
	// | nvd        | 2       | vunnel@0.29.0 | 2025-01-08 01:32:55.179881+00:00 | xxh64:0a160d2b53dd0208 |
	// | alpine     | 1       | vunnel@0.29.0 | 2025-01-08 01:31:28.824872+00:00 | xxh64:30c5b7b8efa0c087 |
	// | amazon     | 1       | vunnel@0.29.0 | 2025-01-08 01:31:28.837469+00:00 | xxh64:7d90b3fa66b183bc |
	// | chainguard | 1       | vunnel@0.29.0 | 2025-01-08 01:31:26.969865+00:00 | xxh64:25a82fa97ac9e077 |
	// | debian     | 1       | vunnel@0.29.0 | 2025-01-08 01:31:50.718966+00:00 | xxh64:4b1834b9e4e68987 |
	// | github     | 1       | vunnel@0.29.0 | 2025-01-08 01:31:27.450124+00:00 | xxh64:a3ee6b48d37a0124 |
	// | mariner    | 1       | vunnel@0.29.0 | 2025-01-08 01:32:35.005761+00:00 | xxh64:cb4f5861a1fda0af |
	// | oracle     | 1       | vunnel@0.29.0 | 2025-01-08 01:32:33.696274+00:00 | xxh64:72c0a15731e96ab3 |
	// | rhel       | 1       | vunnel@0.29.0 | 2025-01-08 01:32:32.192345+00:00 | xxh64:abf5d2fd5a26c194 |
	// | sles       | 1       | vunnel@0.29.0 | 2025-01-08 01:32:42.988937+00:00 | xxh64:8f558f8f28a04489 |
	// | ubuntu     | 3       | vunnel@0.29.0 | 2025-01-08 01:33:25.795537+00:00 | xxh64:97ef8421c0093620 |
	// | wolfi      | 1       | vunnel@0.29.0 | 2025-01-08 01:32:58.571417+00:00 | xxh64:f294f3474d35b1a9 |
	// +------------+---------+---------------+----------------------------------+------------------------+

	// the expected results should mimic what is found as v5 namespace values:
	// +--------------------------------------+
	// | namespace                            |
	// +--------------------------------------+
	// | nvd:cpe                              |
	// | github:language:javascript           |
	// | ubuntu:distro:ubuntu:14.04           |
	// | ubuntu:distro:ubuntu:16.04           |
	// | ubuntu:distro:ubuntu:18.04           |
	// | ubuntu:distro:ubuntu:20.04           |
	// | ubuntu:distro:ubuntu:22.04           |
	// | ubuntu:distro:ubuntu:22.10           |
	// | ubuntu:distro:ubuntu:23.04           |
	// | ubuntu:distro:ubuntu:23.10           |
	// | ubuntu:distro:ubuntu:24.10           |
	// | debian:distro:debian:8               |
	// | debian:distro:debian:9               |
	// | ubuntu:distro:ubuntu:12.04           |
	// | ubuntu:distro:ubuntu:15.04           |
	// | sles:distro:sles:15                  |
	// | sles:distro:sles:15.1                |
	// | sles:distro:sles:15.2                |
	// | sles:distro:sles:15.3                |
	// | sles:distro:sles:15.4                |
	// | sles:distro:sles:15.5                |
	// | sles:distro:sles:15.6                |
	// | amazon:distro:amazonlinux:2          |
	// | debian:distro:debian:10              |
	// | debian:distro:debian:11              |
	// | debian:distro:debian:12              |
	// | debian:distro:debian:unstable        |
	// | oracle:distro:oraclelinux:6          |
	// | oracle:distro:oraclelinux:7          |
	// | oracle:distro:oraclelinux:8          |
	// | oracle:distro:oraclelinux:9          |
	// | redhat:distro:redhat:6               |
	// | redhat:distro:redhat:7               |
	// | redhat:distro:redhat:8               |
	// | redhat:distro:redhat:9               |
	// | ubuntu:distro:ubuntu:12.10           |
	// | ubuntu:distro:ubuntu:13.04           |
	// | ubuntu:distro:ubuntu:14.10           |
	// | ubuntu:distro:ubuntu:15.10           |
	// | ubuntu:distro:ubuntu:16.10           |
	// | ubuntu:distro:ubuntu:17.04           |
	// | ubuntu:distro:ubuntu:17.10           |
	// | ubuntu:distro:ubuntu:18.10           |
	// | ubuntu:distro:ubuntu:19.04           |
	// | ubuntu:distro:ubuntu:19.10           |
	// | ubuntu:distro:ubuntu:20.10           |
	// | ubuntu:distro:ubuntu:21.04           |
	// | ubuntu:distro:ubuntu:21.10           |
	// | ubuntu:distro:ubuntu:24.04           |
	// | github:language:php                  |
	// | debian:distro:debian:13              |
	// | debian:distro:debian:7               |
	// | redhat:distro:redhat:5               |
	// | sles:distro:sles:11.1                |
	// | sles:distro:sles:11.3                |
	// | sles:distro:sles:11.4                |
	// | sles:distro:sles:11.2                |
	// | sles:distro:sles:12                  |
	// | sles:distro:sles:12.1                |
	// | sles:distro:sles:12.2                |
	// | sles:distro:sles:12.3                |
	// | sles:distro:sles:12.4                |
	// | sles:distro:sles:12.5                |
	// | chainguard:distro:chainguard:rolling |
	// | wolfi:distro:wolfi:rolling           |
	// | github:language:go                   |
	// | alpine:distro:alpine:3.20            |
	// | alpine:distro:alpine:3.21            |
	// | alpine:distro:alpine:edge            |
	// | github:language:rust                 |
	// | github:language:python               |
	// | sles:distro:sles:11                  |
	// | oracle:distro:oraclelinux:5          |
	// | github:language:ruby                 |
	// | github:language:dotnet               |
	// | alpine:distro:alpine:3.12            |
	// | alpine:distro:alpine:3.13            |
	// | alpine:distro:alpine:3.14            |
	// | alpine:distro:alpine:3.15            |
	// | alpine:distro:alpine:3.16            |
	// | alpine:distro:alpine:3.17            |
	// | alpine:distro:alpine:3.18            |
	// | alpine:distro:alpine:3.19            |
	// | mariner:distro:mariner:2.0           |
	// | github:language:java                 |
	// | github:language:dart                 |
	// | amazon:distro:amazonlinux:2023       |
	// | alpine:distro:alpine:3.10            |
	// | alpine:distro:alpine:3.11            |
	// | alpine:distro:alpine:3.4             |
	// | alpine:distro:alpine:3.5             |
	// | alpine:distro:alpine:3.7             |
	// | alpine:distro:alpine:3.8             |
	// | alpine:distro:alpine:3.9             |
	// | mariner:distro:azurelinux:3.0        |
	// | mariner:distro:mariner:1.0           |
	// | alpine:distro:alpine:3.3             |
	// | alpine:distro:alpine:3.6             |
	// | amazon:distro:amazonlinux:2022       |
	// | alpine:distro:alpine:3.2             |
	// | github:language:swift                |
	// +--------------------------------------+

	type testCase struct {
		name      string
		provider  string // from Providers.id
		ecosystem string // only used when provider is "github"
		osName    string // only used for OS-based providers
		osVersion string // only used for OS-based providers
		expected  string
	}

	tests := []testCase{
		// NVD
		{
			name:     "nvd provider",
			provider: "nvd",
			expected: "nvd:cpe",
		},

		// GitHub ecosystem tests
		{
			name:      "github golang direct",
			provider:  "github",
			ecosystem: "golang",
			expected:  "github:language:go",
		},
		{
			name:      "github go-module ecosystem",
			provider:  "github",
			ecosystem: "go-module",
			expected:  "github:language:go",
		},
		{
			name:      "github composer ecosystem",
			provider:  "github",
			ecosystem: "composer",
			expected:  "github:language:php",
		},
		{
			name:      "github php-composer ecosystem",
			provider:  "github",
			ecosystem: "php-composer",
			expected:  "github:language:php",
		},
		{
			name:      "github cargo ecosystem",
			provider:  "github",
			ecosystem: "cargo",
			expected:  "github:language:rust",
		},
		{
			name:      "github rust-crate ecosystem",
			provider:  "github",
			ecosystem: "rust-crate",
			expected:  "github:language:rust",
		},
		{
			name:      "github pub ecosystem",
			provider:  "github",
			ecosystem: "pub",
			expected:  "github:language:dart",
		},
		{
			name:      "github dart-pub ecosystem",
			provider:  "github",
			ecosystem: "dart-pub",
			expected:  "github:language:dart",
		},
		{
			name:      "github nuget ecosystem",
			provider:  "github",
			ecosystem: "nuget",
			expected:  "github:language:dotnet",
		},
		{
			name:      "github maven ecosystem",
			provider:  "github",
			ecosystem: "maven",
			expected:  "github:language:java",
		},
		{
			name:      "github swifturl ecosystem",
			provider:  "github",
			ecosystem: "swifturl",
			expected:  "github:language:swift",
		},
		{
			name:      "github npm ecosystem",
			provider:  "github",
			ecosystem: "npm",
			expected:  "github:language:javascript",
		},
		{
			name:      "github node ecosystem",
			provider:  "github",
			ecosystem: "node",
			expected:  "github:language:javascript",
		},
		{
			name:      "github pypi ecosystem",
			provider:  "github",
			ecosystem: "pypi",
			expected:  "github:language:python",
		},
		{
			name:      "github pip ecosystem",
			provider:  "github",
			ecosystem: "pip",
			expected:  "github:language:python",
		},
		{
			name:      "github rubygems ecosystem",
			provider:  "github",
			ecosystem: "rubygems",
			expected:  "github:language:ruby",
		},
		{
			name:      "github gem ecosystem",
			provider:  "github",
			ecosystem: "gem",
			expected:  "github:language:ruby",
		},

		// OS Distribution tests
		{
			name:      "ubuntu distribution",
			provider:  "ubuntu",
			osName:    "ubuntu",
			osVersion: "22.04",
			expected:  "ubuntu:distro:ubuntu:22.04",
		},
		{
			name:      "redhat distribution",
			provider:  "rhel",
			osName:    "redhat",
			osVersion: "8",
			expected:  "redhat:distro:redhat:8",
		},
		{
			name:      "debian distribution",
			provider:  "debian",
			osName:    "debian",
			osVersion: "11",
			expected:  "debian:distro:debian:11",
		},
		{
			name:      "sles distribution",
			provider:  "sles",
			osName:    "sles",
			osVersion: "15.5",
			expected:  "sles:distro:sles:15.5",
		},
		{
			name:      "alpine distribution",
			provider:  "alpine",
			osName:    "alpine",
			osVersion: "3.18",
			expected:  "alpine:distro:alpine:3.18",
		},
		{
			name:      "chainguard distribution",
			provider:  "chainguard",
			osName:    "chainguard",
			osVersion: "rolling",
			expected:  "chainguard:distro:chainguard:rolling",
		},
		{
			name:      "wolfi distribution",
			provider:  "wolfi",
			osName:    "wolfi",
			osVersion: "rolling",
			expected:  "wolfi:distro:wolfi:rolling",
		},
		{
			name:      "amazon linux distribution",
			provider:  "amazon",
			osName:    "amazon",
			osVersion: "2023",
			expected:  "amazon:distro:amazonlinux:2023",
		},
		{
			name:      "mariner regular version",
			provider:  "mariner",
			osName:    "mariner",
			osVersion: "2.0",
			expected:  "mariner:distro:mariner:2.0",
		},
		{
			name:      "mariner azure version",
			provider:  "mariner",
			osName:    "mariner",
			osVersion: "3.0",
			expected:  "mariner:distro:azurelinux:3.0",
		},
		{
			name:      "oracle linux distribution",
			provider:  "oracle",
			osName:    "oracle",
			osVersion: "8",
			expected:  "oracle:distro:oraclelinux:8",
		},

		// Version truncation tests
		{
			name:      "rhel with minor version",
			provider:  "rhel",
			osName:    "redhat",
			osVersion: "8.6",
			expected:  "redhat:distro:redhat:8",
		},
		{
			name:      "rhel with patch version",
			provider:  "rhel",
			osName:    "redhat",
			osVersion: "9.2.1",
			expected:  "redhat:distro:redhat:9",
		},
		{
			name:      "oracle with minor version",
			provider:  "oracle",
			osName:    "oracle",
			osVersion: "8.7",
			expected:  "oracle:distro:oraclelinux:8",
		},
		{
			name:      "oracle with patch version",
			provider:  "oracle",
			osName:    "oracle",
			osVersion: "9.3.1",
			expected:  "oracle:distro:oraclelinux:9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := dbsearch.AffectedPackage{
				Vulnerability: dbsearch.VulnerabilityInfo{
					Provider: tt.provider,
				},
			}

			// Add OS info for OS-based providers
			if tt.osName != "" {
				input.AffectedPackageInfo.OS = &dbsearch.OperatingSystem{
					Name:    tt.osName,
					Version: tt.osVersion,
				}
			}

			// Add package info for GitHub provider
			if tt.provider == "github" {
				input.AffectedPackageInfo.Package = &dbsearch.Package{
					Ecosystem: tt.ecosystem,
				}
			}

			result := v5Namespace(input)
			assert.Equal(t, tt.expected, result)
		})
	}
}