package authsvc

import "context"

// EmailSender sends transactional auth emails. Inject a real provider (SES, SendGrid, etc.).
type EmailSender interface {
	SendEmailVerification(ctx context.Context, toEmail, plaintextToken string) error
	SendPasswordReset(ctx context.Context, toEmail, plaintextToken string) error
}

// OTPSender delivers one-time codes for phone verification. Inject SMS provider (Twilio, etc.).
type OTPSender interface {
	SendOTP(ctx context.Context, e164Phone, code string) error
}

type Option func(*Service)

func WithEmailSender(sender EmailSender) Option {
	return func(s *Service) {
		s.emailSender = sender
	}
}

func WithOTPSender(sender OTPSender) Option {
	return func(s *Service) {
		s.otpSender = sender
	}
}
