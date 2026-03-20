package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

type stubRepository struct {
	upserted            []SubscriptionUpsertParams
	subscriptionByID    map[string]StoredSubscription
	checkoutSessionByID map[string]StoredSubscription
}

func (r *stubRepository) UpsertSubscription(_ context.Context, params SubscriptionUpsertParams) (StoredSubscription, error) {
	r.upserted = append(r.upserted, params)
	return StoredSubscription{
		ID:                    "sub-row-1",
		OrganizationID:        params.OrganizationID,
		Provider:              params.Provider,
		StripeCustomerID:      params.StripeCustomerID,
		StripeSubscriptionID:  params.StripeSubscriptionID,
		StripeCheckoutSession: params.StripeCheckoutSession,
		Status:                params.Status,
	}, nil
}

func (r *stubRepository) GetByStripeSubscriptionID(_ context.Context, stripeSubscriptionID string) (StoredSubscription, error) {
	if stored, ok := r.subscriptionByID[stripeSubscriptionID]; ok {
		return stored, nil
	}
	return StoredSubscription{}, ErrSubscriptionNotFound
}

func (r *stubRepository) GetByStripeCheckoutSession(_ context.Context, stripeCheckoutSessionID string) (StoredSubscription, error) {
	if stored, ok := r.checkoutSessionByID[stripeCheckoutSessionID]; ok {
		return stored, nil
	}
	return StoredSubscription{}, ErrSubscriptionNotFound
}

type stubDoer struct {
	response *http.Response
	err      error
}

func (d stubDoer) Do(_ *http.Request) (*http.Response, error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.response, nil
}

func TestCheckoutRequiresConfiguration(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo, Config{})

	_, err := svc.Checkout(context.Background(), "org-1")
	if !errors.Is(err, ErrBillingUnavailable) {
		t.Fatalf("expected ErrBillingUnavailable, got %v", err)
	}
}

func TestCheckoutCreatesPendingSubscription(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo, Config{
		StripeSecretKey: "sk_test_123",
		StripePriceID:   "price_test_123",
		AppBaseURL:      "https://app.example.com",
		HTTPClient: stubDoer{
			response: newJSONResponse(http.StatusOK, `{"id":"cs_test_123","url":"https://checkout.stripe.test/session/cs_test_123","customer":"cus_test_123","status":"open"}`),
		},
	})

	response, err := svc.Checkout(context.Background(), "org-1")
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if response.SessionID != "cs_test_123" {
		t.Fatalf("expected checkout session id, got %+v", response)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected one upsert, got %d", len(repo.upserted))
	}
	if repo.upserted[0].OrganizationID != "org-1" {
		t.Fatalf("expected organization id org-1, got %q", repo.upserted[0].OrganizationID)
	}
}

func TestProcessWebhookRejectsInvalidSignature(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo, Config{StripeWebhookSecret: "whsec_test"})

	_, err := svc.ProcessWebhook(context.Background(), []byte(`{"id":"evt_test","type":"customer.subscription.updated"}`), "t=1,v1=bad")
	if !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Fatalf("expected ErrInvalidWebhookSignature, got %v", err)
	}
}

func TestProcessWebhookUpdatesSubscription(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo, Config{StripeWebhookSecret: "whsec_test"})

	payload := []byte(`{"id":"evt_1","type":"customer.subscription.updated","data":{"object":{"id":"sub_123","customer":"cus_123","metadata":{"organization_id":"org_123"},"status":"active","cancel_at_period_end":false,"current_period_start":1735689600,"current_period_end":1738368000}}}`)
	signature := signTestPayload(payload, "whsec_test", time.Now())

	received, err := svc.ProcessWebhook(context.Background(), payload, signature)
	if err != nil {
		t.Fatalf("process webhook: %v", err)
	}
	if !received {
		t.Fatalf("expected received=true")
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected one upsert, got %d", len(repo.upserted))
	}
	if repo.upserted[0].OrganizationID != "org_123" {
		t.Fatalf("expected org_123, got %q", repo.upserted[0].OrganizationID)
	}
	if repo.upserted[0].StripeSubscriptionID != "sub_123" {
		t.Fatalf("expected sub_123, got %q", repo.upserted[0].StripeSubscriptionID)
	}
}

func TestProcessWebhookFallsBackToStoredCheckoutOrganization(t *testing.T) {
	repo := &stubRepository{
		checkoutSessionByID: map[string]StoredSubscription{
			"cs_fallback": {
				OrganizationID: "org_fallback",
			},
		},
	}
	svc := NewService(repo, Config{StripeWebhookSecret: "whsec_test"})

	payload := []byte(`{"id":"evt_2","type":"checkout.session.completed","data":{"object":{"id":"cs_fallback","customer":"cus_123","subscription":"sub_123","status":"complete","payment_status":"paid"}}}`)
	signature := signTestPayload(payload, "whsec_test", time.Now())

	received, err := svc.ProcessWebhook(context.Background(), payload, signature)
	if err != nil {
		t.Fatalf("process webhook: %v", err)
	}
	if !received {
		t.Fatalf("expected received=true")
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected one upsert, got %d", len(repo.upserted))
	}
	if repo.upserted[0].OrganizationID != "org_fallback" {
		t.Fatalf("expected fallback org, got %q", repo.upserted[0].OrganizationID)
	}
	if repo.upserted[0].Status != "paid" {
		t.Fatalf("expected paid status, got %q", repo.upserted[0].Status)
	}
}

func signTestPayload(payload []byte, secret string, now time.Time) string {
	timestamp := strconv.FormatInt(now.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "." + string(payload)))
	signature := hex.EncodeToString(mac.Sum(nil))
	return strings.Join([]string{"t=" + timestamp, "v1=" + signature}, ",")
}
