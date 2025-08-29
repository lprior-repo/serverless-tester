package dsl

import (
	"fmt"
	"time"
)

// SESBuilder provides a fluent interface for SES operations
type SESBuilder struct {
	ctx *TestContext
}

// Email creates an email builder for sending emails
func (sb *SESBuilder) Email() *SESEmailBuilder {
	return &SESEmailBuilder{
		ctx:    sb.ctx,
		config: &EmailConfig{},
	}
}

// Identity creates an identity builder for managing identities
func (sb *SESBuilder) Identity(email string) *SESIdentityBuilder {
	return &SESIdentityBuilder{
		ctx:      sb.ctx,
		identity: email,
		config:   &IdentityConfig{},
	}
}

// Template creates a template builder for managing email templates
func (sb *SESBuilder) Template(name string) *SESTemplateBuilder {
	return &SESTemplateBuilder{
		ctx:          sb.ctx,
		templateName: name,
		config:       &TemplateConfig{},
	}
}

// ConfigurationSet creates a configuration set builder
func (sb *SESBuilder) ConfigurationSet(name string) *SESConfigSetBuilder {
	return &SESConfigSetBuilder{
		ctx:    sb.ctx,
		name:   name,
		config: &ConfigSetConfig{},
	}
}

// SESEmailBuilder builds email sending operations
type SESEmailBuilder struct {
	ctx    *TestContext
	config *EmailConfig
}

type EmailConfig struct {
	From            string
	To              []string
	CC              []string
	BCC             []string
	Subject         string
	Body            string
	BodyHTML        string
	ReplyTo         []string
	ConfigSet       string
	Tags            map[string]string
	TemplateData    map[string]interface{}
	TemplateName    string
}

// From sets the sender email address
func (seb *SESEmailBuilder) From(email string) *SESEmailBuilder {
	seb.config.From = email
	return seb
}

// To adds recipient email addresses
func (seb *SESEmailBuilder) To(emails ...string) *SESEmailBuilder {
	seb.config.To = append(seb.config.To, emails...)
	return seb
}

// CC adds CC email addresses
func (seb *SESEmailBuilder) CC(emails ...string) *SESEmailBuilder {
	seb.config.CC = append(seb.config.CC, emails...)
	return seb
}

// BCC adds BCC email addresses
func (seb *SESEmailBuilder) BCC(emails ...string) *SESEmailBuilder {
	seb.config.BCC = append(seb.config.BCC, emails...)
	return seb
}

// WithSubject sets the email subject
func (seb *SESEmailBuilder) WithSubject(subject string) *SESEmailBuilder {
	seb.config.Subject = subject
	return seb
}

// WithTextBody sets the plain text body
func (seb *SESEmailBuilder) WithTextBody(body string) *SESEmailBuilder {
	seb.config.Body = body
	return seb
}

// WithHTMLBody sets the HTML body
func (seb *SESEmailBuilder) WithHTMLBody(html string) *SESEmailBuilder {
	seb.config.BodyHTML = html
	return seb
}

// WithTemplate uses an email template
func (seb *SESEmailBuilder) WithTemplate(templateName string, data map[string]interface{}) *SESEmailBuilder {
	seb.config.TemplateName = templateName
	seb.config.TemplateData = data
	return seb
}

// WithTags adds email tags
func (seb *SESEmailBuilder) WithTags(tags map[string]string) *SESEmailBuilder {
	seb.config.Tags = tags
	return seb
}

// Send sends the email
func (seb *SESEmailBuilder) Send() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"MessageId": "0000014a-f896-4c4b-b19d-6c284d6ad55c-000000",
			"Recipients": len(seb.config.To) + len(seb.config.CC) + len(seb.config.BCC),
		},
		Metadata: map[string]interface{}{
			"from":    seb.config.From,
			"subject": seb.config.Subject,
			"to_count": len(seb.config.To),
		},
		ctx: seb.ctx,
	}
}

// SESIdentityBuilder builds identity management operations
type SESIdentityBuilder struct {
	ctx      *TestContext
	identity string
	config   *IdentityConfig
}

type IdentityConfig struct {
	NotificationTopic string
	ForwardingEnabled bool
}

// Verify verifies the identity
func (sib *SESIdentityBuilder) Verify() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Verification email sent to %s", sib.identity),
		Metadata: map[string]interface{}{
			"identity": sib.identity,
			"status":   "Pending",
		},
		ctx: sib.ctx,
	}
}

// GetVerificationStatus gets the verification status
func (sib *SESIdentityBuilder) GetVerificationStatus() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"VerificationStatus": "Success",
			"Identity":           sib.identity,
		},
		ctx: sib.ctx,
	}
}

// Delete deletes the identity
func (sib *SESIdentityBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Identity %s deleted", sib.identity),
		ctx:     sib.ctx,
	}
}

// SetNotificationTopic sets the notification topic for bounces/complaints
func (sib *SESIdentityBuilder) SetNotificationTopic(topic string) *SESIdentityBuilder {
	sib.config.NotificationTopic = topic
	return sib
}

// EnableForwarding enables email forwarding
func (sib *SESIdentityBuilder) EnableForwarding() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Forwarding enabled for %s", sib.identity),
		ctx:     sib.ctx,
	}
}

// SESTemplateBuilder builds template management operations
type SESTemplateBuilder struct {
	ctx          *TestContext
	templateName string
	config       *TemplateConfig
}

type TemplateConfig struct {
	Subject  string
	TextPart string
	HTMLPart string
}

// WithSubject sets the template subject
func (stb *SESTemplateBuilder) WithSubject(subject string) *SESTemplateBuilder {
	stb.config.Subject = subject
	return stb
}

// WithTextPart sets the template text part
func (stb *SESTemplateBuilder) WithTextPart(text string) *SESTemplateBuilder {
	stb.config.TextPart = text
	return stb
}

// WithHTMLPart sets the template HTML part
func (stb *SESTemplateBuilder) WithHTMLPart(html string) *SESTemplateBuilder {
	stb.config.HTMLPart = html
	return stb
}

// Create creates the template
func (stb *SESTemplateBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"TemplateName": stb.templateName,
			"Status":       "Created",
		},
		ctx: stb.ctx,
	}
}

// Update updates the template
func (stb *SESTemplateBuilder) Update() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Template %s updated", stb.templateName),
		ctx:     stb.ctx,
	}
}

// Delete deletes the template
func (stb *SESTemplateBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Template %s deleted", stb.templateName),
		ctx:     stb.ctx,
	}
}

// GetTemplate gets template details
func (stb *SESTemplateBuilder) GetTemplate() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"TemplateName": stb.templateName,
			"Subject":      stb.config.Subject,
			"TextPart":     stb.config.TextPart,
			"HtmlPart":     stb.config.HTMLPart,
		},
		ctx: stb.ctx,
	}
}

// SESConfigSetBuilder builds configuration set operations
type SESConfigSetBuilder struct {
	ctx    *TestContext
	name   string
	config *ConfigSetConfig
}

type ConfigSetConfig struct {
	EventDestinations []EventDestination
	ReputationTracking bool
	DeliveryOptions   DeliveryOptions
}

type EventDestination struct {
	Name        string
	Enabled     bool
	EventTypes  []string
	SNSTopicARN string
}

type DeliveryOptions struct {
	TLSPolicy string
}

// WithEventDestination adds an event destination
func (scb *SESConfigSetBuilder) WithEventDestination(name string, snsTopicARN string, eventTypes []string) *SESConfigSetBuilder {
	scb.config.EventDestinations = append(scb.config.EventDestinations, EventDestination{
		Name:        name,
		Enabled:     true,
		EventTypes:  eventTypes,
		SNSTopicARN: snsTopicARN,
	})
	return scb
}

// WithReputationTracking enables reputation tracking
func (scb *SESConfigSetBuilder) WithReputationTracking() *SESConfigSetBuilder {
	scb.config.ReputationTracking = true
	return scb
}

// Create creates the configuration set
func (scb *SESConfigSetBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"ConfigurationSetName": scb.name,
			"Status":               "Active",
		},
		ctx: scb.ctx,
	}
}

// Delete deletes the configuration set
func (scb *SESConfigSetBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Configuration set %s deleted", scb.name),
		ctx:     scb.ctx,
	}
}

// GetSendQuota gets the sending quota
func (sb *SESBuilder) GetSendQuota() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"Max24HourSend":   200.0,
			"MaxSendRate":     1.0,
			"SentLast24Hours": 0.0,
		},
		ctx: sb.ctx,
	}
}

// GetSendStatistics gets send statistics
func (sb *SESBuilder) GetSendStatistics() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: []map[string]interface{}{
			{
				"Timestamp":       time.Now().Add(-time.Hour),
				"DeliveryAttempts": 10,
				"Bounces":         0,
				"Complaints":      0,
				"Rejects":         0,
			},
		},
		ctx: sb.ctx,
	}
}