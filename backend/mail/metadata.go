package mail

import "fmt"

// Metadata provides common information for email templates,
// such as contact details and URLs for footers.
type Metadata struct {
	// Organization/sender identification
	OrganisationName     string
	HelpdeskEmailAddress string

	// Logo for HTML emails (use GetCID for inline images)
	LogoURL string
	LogoAlt string

	// Contact information
	PhoneNumber  string
	PhoneLabel   string
	WebsiteURL   string
	WebsiteLabel string

	// Legal links
	ImprintURL            string
	PrivacyPolicyURL      string
	TermsAndConditionsURL string
	ContactURL            string
}

// NewMetadata creates Metadata from mail configuration.
// Override fields as needed for specific use cases.
func NewMetadata(config Config) Metadata {
	return Metadata{
		OrganisationName:     config.SenderName,
		HelpdeskEmailAddress: config.HelpdeskEmailAddress,
	}
}

// GetCID returns a Content-ID reference for embedding images in HTML emails.
// Use this with EmbedFromEmbedFS to reference embedded assets.
func GetCID(filename string) string {
	if filename == "" {
		return ""
	}

	return fmt.Sprintf("cid:%s", filename)
}
