package zendesk

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ParseZendeskURL extracts the subdomain and ticket_id from a Zendesk ticket URL
func ParseZendeskURL(zendeskURL string) (string, string, error) {
	if !strings.HasPrefix(zendeskURL, "http") {
		zendeskURL = "http://" + zendeskURL
	}
	// Parse the URL
	u, err := url.Parse(zendeskURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %v", err)
	}

	// Extract the subdomain from the host (e.g., "temporalsupport" from "temporalsupport.zendesk.com")
	hostParts := strings.Split(u.Hostname(), ".")
	if len(hostParts) < 2 {
		return "", "", fmt.Errorf("invalid Zendesk subdomain in URL: %v", zendeskURL)
	}
	subdomain := hostParts[0]

	// Use regex to extract the ticket ID from the path
	re := regexp.MustCompile(`/tickets/(\d+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) < 2 {
		return "", "", fmt.Errorf("ticket ID not found in URL path: %v", zendeskURL)
	}
	ticketID := matches[1]

	return subdomain, ticketID, nil
}
