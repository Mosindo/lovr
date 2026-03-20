package billing

import (
	"encoding/json"
	"time"
)

type Subscription struct {
	ID                    string
	OrganizationID        string
	Provider              string
	StripeCustomerID      string
	StripeSubscriptionID  string
	StripeCheckoutSession string
	Status                string
	CurrentPeriodStart    *time.Time
	CurrentPeriodEnd      *time.Time
	CancelAtPeriodEnd     bool
	CanceledAt            *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CheckoutResponse struct {
	SessionID      string `json:"sessionId"`
	CheckoutURL    string `json:"checkoutUrl"`
	OrganizationID string `json:"organizationId"`
	Status         string `json:"status"`
}

type WebhookResponse struct {
	Received bool `json:"received"`
}

type stripeCheckoutSessionResponse struct {
	ID               string            `json:"id"`
	URL              string            `json:"url"`
	Customer         string            `json:"customer"`
	Subscription     string            `json:"subscription"`
	Status           string            `json:"status"`
	PaymentStatus    string            `json:"payment_status"`
	ClientRefID      string            `json:"client_reference_id"`
	Metadata         map[string]string `json:"metadata"`
	CustomerEmail    string            `json:"customer_email"`
	ExpiresAtUnix    int64             `json:"expires_at"`
	SubscriptionData struct {
		Metadata map[string]string `json:"metadata"`
	} `json:"subscription_data"`
}

type stripeEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

type stripeSubscriptionObject struct {
	ID                 string            `json:"id"`
	Customer           string            `json:"customer"`
	Metadata           map[string]string `json:"metadata"`
	Status             string            `json:"status"`
	CancelAtPeriodEnd  bool              `json:"cancel_at_period_end"`
	CurrentPeriodStart int64             `json:"current_period_start"`
	CurrentPeriodEnd   int64             `json:"current_period_end"`
	CanceledAtUnix     int64             `json:"canceled_at"`
}
