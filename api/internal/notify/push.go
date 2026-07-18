package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Notifier sends push notifications to registered on-call devices via FCM.
// Kept as a thin HTTP client wrapper — no Firebase SDK dependency needed
// for a single legacy HTTP v1 send call.
type Notifier struct {
	ServerKey string
	Pool      *pgxpool.Pool
	client    *http.Client
}

func NewNotifier(serverKey string, pool *pgxpool.Pool) *Notifier {
	return &Notifier{
		ServerKey: serverKey,
		Pool:      pool,
		client:    &http.Client{},
	}
}

type fcmPayload struct {
	To           string            `json:"to"`
	Notification fcmNotification   `json:"notification"`
	Data         map[string]string `json:"data"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// NotifyRiskyPlan pushes an alert to every registered device when a plan
// scan is scored medium/high risk and needs a human decision.
func (n *Notifier) NotifyRiskyPlan(ctx context.Context, scanID, repoName, riskLevel string) error {
	if n.ServerKey == "" {
		// No key configured (e.g. local dev without FCM set up) — skip
		// silently rather than failing the whole scan flow.
		return nil
	}

	rows, err := n.Pool.Query(ctx, `SELECT device_token FROM device_tokens`)
	if err != nil {
		return fmt.Errorf("failed to fetch device tokens: %w", err)
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			continue
		}
		tokens = append(tokens, token)
	}

	if len(tokens) == 0 {
		return nil // nobody registered yet — nothing to do
	}

	for _, token := range tokens {
		payload := fcmPayload{
			To: token,
			Notification: fcmNotification{
				Title: fmt.Sprintf("%s risk plan needs approval", riskLevel),
				Body:  fmt.Sprintf("%s has a %s-risk Terraform change awaiting your decision.", repoName, riskLevel),
			},
			Data: map[string]string{
				"scan_id": scanID,
				"type":    "plan_approval",
			},
		}

		body, err := json.Marshal(payload)
		if err != nil {
			continue
		}

		req, err := http.NewRequestWithContext(
			ctx, http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(body),
		)
		if err != nil {
			continue
		}
		req.Header.Set("Authorization", "key="+n.ServerKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := n.client.Do(req)
		if err != nil {
			// Log-and-continue: one failed push shouldn't block others.
			continue
		}
		resp.Body.Close()
	}

	return nil
}
