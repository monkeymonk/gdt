package metadata

import "strings"

// FindChecksum parses SHA512-SUMS.txt content and returns the checksum
// for the given artifact name. Returns empty string if not found.
func FindChecksum(checksumContent string, artifactName string) string {
	if checksumContent == "" || artifactName == "" {
		return ""
	}
	for _, line := range strings.Split(checksumContent, "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == artifactName {
			return parts[0]
		}
	}
	return ""
}
