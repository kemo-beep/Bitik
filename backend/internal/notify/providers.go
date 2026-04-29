package notify

import "context"

type ProviderType string

const (
	ProviderInApp ProviderType = "in_app"
	ProviderEmail ProviderType = "email"
	ProviderSMS   ProviderType = "sms"
	ProviderPush  ProviderType = "push"
)

type Provider interface {
	Type() ProviderType
	Send(ctx context.Context, evt Event) error
}

