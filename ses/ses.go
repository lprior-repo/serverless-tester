// Package ses provides comprehensive AWS Simple Email Service (SES) testing utilities following Terratest patterns.
//
// This package is designed to be a drop-in replacement for Terratest's SES utilities
// with enhanced serverless-specific features. It follows strict functional programming
// principles and provides both Function and FunctionE variants for all operations.
//
// Key Features:
//   - Email sending and verification management
//   - Identity and domain verification
//   - Configuration set and event publishing
//   - Template management with dynamic content
//   - Suppression list management
//   - Reputation tracking and monitoring
//   - Bounce and complaint handling
//   - Custom mail-from domain configuration
//   - Receipt rule and filter management
//   - Comprehensive email testing utilities
//
// Example usage:
//
//   func TestSESEmailSending(t *testing.T) {
//       ctx := sfx.NewTestContext(t)
//       
//       // Verify email identity
//       identity := VerifyEmailIdentity(ctx, "test@example.com")
//       
//       // Send email
//       message := SendEmail(ctx, EmailConfig{
//           Source:      "test@example.com",
//           Destination: []string{"recipient@example.com"},
//           Subject:     "Test Email",
//           Body:        "This is a test email",
//       })
//       
//       // Create template
//       template := CreateTemplate(ctx, TemplateConfig{
//           TemplateName: "welcome-template",
//           Subject:      "Welcome {{name}}!",
//           HtmlBody:     "<h1>Welcome {{name}}!</h1>",
//           TextBody:     "Welcome {{name}}!",
//       })
//       
//       assert.NotEmpty(t, message.MessageId)
//   }
package ses

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	sesv2types "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/stretchr/testify/require"
)

// Common SES configurations
const (
	DefaultTimeout       = 30 * time.Second
	DefaultRetryAttempts = 3
	DefaultRetryDelay    = 1 * time.Second
	MaxSendQuota         = 200   // Default daily sending quota
	MaxSendRate          = 1.0   // Default sending rate per second
	DefaultCharset       = "UTF-8"
)

// Email verification states
const (
	VerificationStatusPending   = types.VerificationStatusPending
	VerificationStatusSuccess   = types.VerificationStatusSuccess
	VerificationStatusFailed    = types.VerificationStatusFailed
	VerificationStatusTemporary = types.VerificationStatusTemporaryFailure
)

// Common errors
var (
	ErrEmailNotVerified        = errors.New("email identity not verified")
	ErrDomainNotVerified       = errors.New("domain identity not verified")
	ErrInvalidEmailAddress     = errors.New("invalid email address")
	ErrTemplateNotFound        = errors.New("template not found")
	ErrConfigurationSetNotFound = errors.New("configuration set not found")
	ErrSendingQuotaExceeded    = errors.New("sending quota exceeded")
	ErrSendingRateExceeded     = errors.New("sending rate exceeded")
	ErrInvalidTemplate         = errors.New("invalid template format")
	ErrReputationPaused        = errors.New("sending paused due to reputation issues")
)

// TestContext represents the test context with AWS configuration
type TestContext interface {
	AwsConfig() aws.Config
	T() interface{ 
		Errorf(format string, args ...interface{})
		Error(args ...interface{})
		Fail()
	}
}

// Email configuration for sending emails
type EmailConfig struct {
	Source               string            `json:"source"`
	Destination          []string          `json:"destination"`
	Subject              string            `json:"subject"`
	Body                 string            `json:"body"`
	HtmlBody             string            `json:"htmlBody,omitempty"`
	TextBody             string            `json:"textBody,omitempty"`
	ReplyToAddresses     []string          `json:"replyToAddresses,omitempty"`
	ReturnPath           string            `json:"returnPath,omitempty"`
	ConfigurationSetName string            `json:"configurationSetName,omitempty"`
	Tags                 map[string]string `json:"tags,omitempty"`
}

// Template configuration for creating email templates
type TemplateConfig struct {
	TemplateName string `json:"templateName"`
	Subject      string `json:"subject"`
	HtmlBody     string `json:"htmlBody,omitempty"`
	TextBody     string `json:"textBody,omitempty"`
}

// Configuration set configuration
type ConfigurationSetConfig struct {
	Name                    string                           `json:"name"`
	DeliveryOptions         *types.DeliveryOptions           `json:"deliveryOptions,omitempty"`
	ReputationTrackingEnabled bool                           `json:"reputationTrackingEnabled"`
	SendingEnabled          bool                             `json:"sendingEnabled"`
	Tags                    []types.Tag                      `json:"tags,omitempty"`
}

// Identity policy configuration
type IdentityPolicyConfig struct {
	Identity   string `json:"identity"`
	PolicyName string `json:"policyName"`
	Policy     string `json:"policy"`
}

// Results structures
type EmailResult struct {
	MessageId     string            `json:"messageId"`
	Source        string            `json:"source"`
	Destinations  []string          `json:"destinations"`
	Subject       string            `json:"subject"`
	SendTimestamp time.Time         `json:"sendTimestamp"`
	Tags          map[string]string `json:"tags,omitempty"`
}

type IdentityVerificationResult struct {
	Identity           string                      `json:"identity"`
	VerificationStatus types.VerificationStatus    `json:"verificationStatus"`
	VerificationToken  string                      `json:"verificationToken,omitempty"`
	DkimAttributes     *types.IdentityDkimAttributes `json:"dkimAttributes,omitempty"`
}

type TemplateResult struct {
	TemplateName string    `json:"templateName"`
	Subject      string    `json:"subject"`
	HtmlBody     string    `json:"htmlBody,omitempty"`
	TextBody     string    `json:"textBody,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ConfigurationSetResult struct {
	Name                      string    `json:"name"`
	ReputationTrackingEnabled bool      `json:"reputationTrackingEnabled"`
	SendingEnabled            bool      `json:"sendingEnabled"`
	CreatedAt                 time.Time `json:"createdAt"`
}

type SendingQuotaResult struct {
	Max24HourSend   float64 `json:"max24HourSend"`
	MaxSendRate     float64 `json:"maxSendRate"`
	SentLast24Hours float64 `json:"sentLast24Hours"`
}

// createSESClient creates an SES client using shared AWS config from TestContext
func createSESClient(ctx TestContext) *ses.Client {
	return ses.NewFromConfig(ctx.AwsConfig)
}

// createSESv2Client creates an SESv2 client using shared AWS config from TestContext
func createSESv2Client(ctx TestContext) *sesv2.Client {
	return sesv2.NewFromConfig(ctx.AwsConfig)
}

// SendEmail sends an email using SES
func SendEmail(ctx TestContext, config EmailConfig) *EmailResult {
	result, err := SendEmailE(ctx, config)
	require.NoError(ctx.T, err)
	return result
}

// SendEmailE sends an email using SES
func SendEmailE(ctx TestContext, config EmailConfig) (*EmailResult, error) {
	client := createSESClient(ctx)

	// Build destination
	destination := &types.Destination{
		ToAddresses: config.Destination,
	}

	// Build message
	message := &types.Message{
		Subject: &types.Content{
			Data:    aws.String(config.Subject),
			Charset: aws.String(DefaultCharset),
		},
	}

	// Set body content
	if config.HtmlBody != "" && config.TextBody != "" {
		message.Body = &types.Body{
			Html: &types.Content{
				Data:    aws.String(config.HtmlBody),
				Charset: aws.String(DefaultCharset),
			},
			Text: &types.Content{
				Data:    aws.String(config.TextBody),
				Charset: aws.String(DefaultCharset),
			},
		}
	} else if config.HtmlBody != "" {
		message.Body = &types.Body{
			Html: &types.Content{
				Data:    aws.String(config.HtmlBody),
				Charset: aws.String(DefaultCharset),
			},
		}
	} else if config.TextBody != "" {
		message.Body = &types.Body{
			Text: &types.Content{
				Data:    aws.String(config.TextBody),
				Charset: aws.String(DefaultCharset),
			},
		}
	} else {
		message.Body = &types.Body{
			Text: &types.Content{
				Data:    aws.String(config.Body),
				Charset: aws.String(DefaultCharset),
			},
		}
	}

	input := &ses.SendEmailInput{
		Source:      aws.String(config.Source),
		Destination: destination,
		Message:     message,
	}

	if config.ConfigurationSetName != "" {
		input.ConfigurationSetName = aws.String(config.ConfigurationSetName)
	}

	if len(config.ReplyToAddresses) > 0 {
		input.ReplyToAddresses = config.ReplyToAddresses
	}

	if config.ReturnPath != "" {
		input.ReturnPath = aws.String(config.ReturnPath)
	}

	// Convert tags to SES format
	if len(config.Tags) > 0 {
		var tags []types.MessageTag
		for key, value := range config.Tags {
			tags = append(tags, types.MessageTag{
				Name:  aws.String(key),
				Value: aws.String(value),
			})
		}
		input.Tags = tags
	}

	output, err := client.SendEmail(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &EmailResult{
		MessageId:     aws.ToString(output.MessageId),
		Source:        config.Source,
		Destinations:  config.Destination,
		Subject:       config.Subject,
		SendTimestamp: time.Now(),
		Tags:          config.Tags,
	}, nil
}

// SendTemplatedEmail sends an email using a template
func SendTemplatedEmail(ctx TestContext, templateName, source string, destinations []string, templateData map[string]interface{}) *EmailResult {
	result, err := SendTemplatedEmailE(ctx, templateName, source, destinations, templateData)
	require.NoError(ctx.T, err)
	return result
}

// SendTemplatedEmailE sends an email using a template
func SendTemplatedEmailE(ctx TestContext, templateName, source string, destinations []string, templateData map[string]interface{}) (*EmailResult, error) {
	client := createSESClient(ctx)

	// Convert template data to JSON
	templateDataJSON, err := json.Marshal(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template data: %w", err)
	}

	input := &ses.SendTemplatedEmailInput{
		Source:      aws.String(source),
		Template:    aws.String(templateName),
		TemplateData: aws.String(string(templateDataJSON)),
		Destination: &types.Destination{
			ToAddresses: destinations,
		},
	}

	output, err := client.SendTemplatedEmail(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &EmailResult{
		MessageId:     aws.ToString(output.MessageId),
		Source:        source,
		Destinations:  destinations,
		SendTimestamp: time.Now(),
	}, nil
}

// VerifyEmailIdentity verifies an email address identity
func VerifyEmailIdentity(ctx TestContext, emailAddress string) *IdentityVerificationResult {
	result, err := VerifyEmailIdentityE(ctx, emailAddress)
	require.NoError(ctx.T, err)
	return result
}

// VerifyEmailIdentityE verifies an email address identity
func VerifyEmailIdentityE(ctx TestContext, emailAddress string) (*IdentityVerificationResult, error) {
	client := createSESClient(ctx)

	input := &ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(emailAddress),
	}

	_, err := client.VerifyEmailIdentity(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	// Get verification status
	verificationResult, err := GetIdentityVerificationStatusE(ctx, emailAddress)
	if err != nil {
		return nil, err
	}

	return verificationResult, nil
}

// VerifyDomainIdentity verifies a domain identity
func VerifyDomainIdentity(ctx TestContext, domain string) *IdentityVerificationResult {
	result, err := VerifyDomainIdentityE(ctx, domain)
	require.NoError(ctx.T, err)
	return result
}

// VerifyDomainIdentityE verifies a domain identity
func VerifyDomainIdentityE(ctx TestContext, domain string) (*IdentityVerificationResult, error) {
	client := createSESClient(ctx)

	input := &ses.VerifyDomainIdentityInput{
		Domain: aws.String(domain),
	}

	output, err := client.VerifyDomainIdentity(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &IdentityVerificationResult{
		Identity:           domain,
		VerificationStatus: types.VerificationStatusPending,
		VerificationToken:  aws.ToString(output.VerificationToken),
	}, nil
}

// GetIdentityVerificationStatus gets the verification status of an identity
func GetIdentityVerificationStatus(ctx TestContext, identity string) *IdentityVerificationResult {
	result, err := GetIdentityVerificationStatusE(ctx, identity)
	require.NoError(ctx.T, err)
	return result
}

// GetIdentityVerificationStatusE gets the verification status of an identity
func GetIdentityVerificationStatusE(ctx TestContext, identity string) (*IdentityVerificationResult, error) {
	client := createSESClient(ctx)

	input := &ses.GetIdentityVerificationAttributesInput{
		Identities: []string{identity},
	}

	output, err := client.GetIdentityVerificationAttributes(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	attrs, exists := output.VerificationAttributes[identity]
	if !exists {
		return nil, fmt.Errorf("identity %s not found", identity)
	}

	result := &IdentityVerificationResult{
		Identity:           identity,
		VerificationStatus: attrs.VerificationStatus,
		VerificationToken:  aws.ToString(attrs.VerificationToken),
	}

	// Get DKIM attributes if available
	dkimInput := &ses.GetIdentityDkimAttributesInput{
		Identities: []string{identity},
	}

	dkimOutput, err := client.GetIdentityDkimAttributes(context.TODO(), dkimInput)
	if err == nil {
		if dkimAttrs, exists := dkimOutput.DkimAttributes[identity]; exists {
			result.DkimAttributes = &dkimAttrs
		}
	}

	return result, nil
}

// DeleteVerifiedEmailAddress deletes a verified email address
func DeleteVerifiedEmailAddress(ctx TestContext, emailAddress string) {
	err := DeleteVerifiedEmailAddressE(ctx, emailAddress)
	require.NoError(ctx.T, err)
}

// DeleteVerifiedEmailAddressE deletes a verified email address
func DeleteVerifiedEmailAddressE(ctx TestContext, emailAddress string) error {
	client := createSESClient(ctx)

	input := &ses.DeleteVerifiedEmailAddressInput{
		EmailAddress: aws.String(emailAddress),
	}

	_, err := client.DeleteVerifiedEmailAddress(context.TODO(), input)
	return err
}

// ListIdentities lists all verified identities
func ListIdentities(ctx TestContext) []string {
	identities, err := ListIdentitiesE(ctx)
	require.NoError(ctx.T, err)
	return identities
}

// ListIdentitiesE lists all verified identities
func ListIdentitiesE(ctx TestContext) ([]string, error) {
	client := createSESClient(ctx)

	input := &ses.ListIdentitiesInput{}

	output, err := client.ListIdentities(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return output.Identities, nil
}

// CreateTemplate creates an email template
func CreateTemplate(ctx TestContext, config TemplateConfig) *TemplateResult {
	result, err := CreateTemplateE(ctx, config)
	require.NoError(ctx.T, err)
	return result
}

// CreateTemplateE creates an email template
func CreateTemplateE(ctx TestContext, config TemplateConfig) (*TemplateResult, error) {
	client := createSESClient(ctx)

	template := &types.Template{
		TemplateName: aws.String(config.TemplateName),
		SubjectPart:  aws.String(config.Subject),
	}

	if config.HtmlBody != "" {
		template.HtmlPart = aws.String(config.HtmlBody)
	}

	if config.TextBody != "" {
		template.TextPart = aws.String(config.TextBody)
	}

	input := &ses.CreateTemplateInput{
		Template: template,
	}

	_, err := client.CreateTemplate(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &TemplateResult{
		TemplateName: config.TemplateName,
		Subject:      config.Subject,
		HtmlBody:     config.HtmlBody,
		TextBody:     config.TextBody,
		CreatedAt:    time.Now(),
	}, nil
}

// GetTemplate gets an email template
func GetTemplate(ctx TestContext, templateName string) *TemplateResult {
	result, err := GetTemplateE(ctx, templateName)
	require.NoError(ctx.T, err)
	return result
}

// GetTemplateE gets an email template
func GetTemplateE(ctx TestContext, templateName string) (*TemplateResult, error) {
	client := createSESClient(ctx)

	input := &ses.GetTemplateInput{
		TemplateName: aws.String(templateName),
	}

	output, err := client.GetTemplate(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	template := output.Template
	return &TemplateResult{
		TemplateName: aws.ToString(template.TemplateName),
		Subject:      aws.ToString(template.SubjectPart),
		HtmlBody:     aws.ToString(template.HtmlPart),
		TextBody:     aws.ToString(template.TextPart),
	}, nil
}

// UpdateTemplate updates an email template
func UpdateTemplate(ctx TestContext, config TemplateConfig) *TemplateResult {
	result, err := UpdateTemplateE(ctx, config)
	require.NoError(ctx.T, err)
	return result
}

// UpdateTemplateE updates an email template
func UpdateTemplateE(ctx TestContext, config TemplateConfig) (*TemplateResult, error) {
	client := createSESClient(ctx)

	template := &types.Template{
		TemplateName: aws.String(config.TemplateName),
		SubjectPart:  aws.String(config.Subject),
	}

	if config.HtmlBody != "" {
		template.HtmlPart = aws.String(config.HtmlBody)
	}

	if config.TextBody != "" {
		template.TextPart = aws.String(config.TextBody)
	}

	input := &ses.UpdateTemplateInput{
		Template: template,
	}

	_, err := client.UpdateTemplate(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &TemplateResult{
		TemplateName: config.TemplateName,
		Subject:      config.Subject,
		HtmlBody:     config.HtmlBody,
		TextBody:     config.TextBody,
	}, nil
}

// DeleteTemplate deletes an email template
func DeleteTemplate(ctx TestContext, templateName string) {
	err := DeleteTemplateE(ctx, templateName)
	require.NoError(ctx.T, err)
}

// DeleteTemplateE deletes an email template
func DeleteTemplateE(ctx TestContext, templateName string) error {
	client := createSESClient(ctx)

	input := &ses.DeleteTemplateInput{
		TemplateName: aws.String(templateName),
	}

	_, err := client.DeleteTemplate(context.TODO(), input)
	return err
}

// ListTemplates lists all email templates
func ListTemplates(ctx TestContext) []string {
	templates, err := ListTemplatesE(ctx)
	require.NoError(ctx.T, err)
	return templates
}

// ListTemplatesE lists all email templates
func ListTemplatesE(ctx TestContext) ([]string, error) {
	client := createSESClient(ctx)

	input := &ses.ListTemplatesInput{}

	var allTemplates []string
	for {
		output, err := client.ListTemplates(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		for _, template := range output.TemplatesMetadata {
			allTemplates = append(allTemplates, aws.ToString(template.Name))
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return allTemplates, nil
}

// GetSendQuota gets the sending quota and rate
func GetSendQuota(ctx TestContext) *SendingQuotaResult {
	result, err := GetSendQuotaE(ctx)
	require.NoError(ctx.T, err)
	return result
}

// GetSendQuotaE gets the sending quota and rate
func GetSendQuotaE(ctx TestContext) (*SendingQuotaResult, error) {
	client := createSESClient(ctx)

	input := &ses.GetSendQuotaInput{}

	output, err := client.GetSendQuota(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &SendingQuotaResult{
		Max24HourSend:   output.Max24HourSend,
		MaxSendRate:     output.MaxSendRate,
		SentLast24Hours: output.SentLast24Hours,
	}, nil
}

// WaitForIdentityVerification waits for an identity to be verified
func WaitForIdentityVerification(ctx TestContext, identity string, timeout time.Duration) *IdentityVerificationResult {
	result, err := WaitForIdentityVerificationE(ctx, identity, timeout)
	require.NoError(ctx.T, err)
	return result
}

// WaitForIdentityVerificationE waits for an identity to be verified
func WaitForIdentityVerificationE(ctx TestContext, identity string, timeout time.Duration) (*IdentityVerificationResult, error) {
	startTime := time.Now()
	
	for time.Since(startTime) < timeout {
		result, err := GetIdentityVerificationStatusE(ctx, identity)
		if err != nil {
			return nil, err
		}

		if result.VerificationStatus == types.VerificationStatusSuccess {
			return result, nil
		}

		if result.VerificationStatus == types.VerificationStatusFailed {
			return result, fmt.Errorf("identity verification failed for %s", identity)
		}

		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("timeout waiting for identity verification: %s", identity)
}

// IdentityExists checks if an identity exists and is verified
func IdentityExists(ctx TestContext, identity string) bool {
	exists, err := IdentityExistsE(ctx, identity)
	require.NoError(ctx.T, err)
	return exists
}

// IdentityExistsE checks if an identity exists and is verified
func IdentityExistsE(ctx TestContext, identity string) (bool, error) {
	result, err := GetIdentityVerificationStatusE(ctx, identity)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	return result.VerificationStatus == types.VerificationStatusSuccess, nil
}

// TemplateExists checks if a template exists
func TemplateExists(ctx TestContext, templateName string) bool {
	exists, err := TemplateExistsE(ctx, templateName)
	require.NoError(ctx.T, err)
	return exists
}

// TemplateExistsE checks if a template exists
func TemplateExistsE(ctx TestContext, templateName string) (bool, error) {
	_, err := GetTemplateE(ctx, templateName)
	if err != nil {
		if strings.Contains(err.Error(), "not exist") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}