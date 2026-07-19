package templates

import "context"

type contextKey string

const csrfTokenKey contextKey = "csrf_token"

// WithCSRFToken stores the request's CSRF token in the context so form templates
// can embed it without threading it through every component signature.
func WithCSRFToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfTokenKey, token)
}

// csrfToken reads the token published by WithCSRFToken; empty if unset.
func csrfToken(ctx context.Context) string {
	token, _ := ctx.Value(csrfTokenKey).(string)
	return token
}
