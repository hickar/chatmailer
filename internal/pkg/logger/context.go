package logger

import (
	"context"
	"log/slog"
)

// ctxKey is a custom struct used for getting/setting log attributes values.
type ctxKey struct{}

// ContextHandler consideres log attributes stored within context.Context keys,
// to be logged by methods like slog.InfoContext.
type ContextHandler struct {
	slog.Handler
}

// NewContextHandler creates new ContextHandler instance
// with provider handler as it's base.
func NewContextHandler(handler slog.Handler) *ContextHandler {
	return &ContextHandler{Handler: handler}
}

// Handle adds contextual attributes to slog.Record entry
// before calling underlying handler.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(ctxKey{}).([]slog.Attr); ok {
		r.AddAttrs(attrs...)
	}

	return h.Handler.Handle(ctx, r)
}

// WithAttrs creates new context.Context value
// with slog attributes stored within it.
func WithAttrs(parent context.Context, attrs ...slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(ctxKey{}).([]slog.Attr); ok {
		v = append(v, attrs...)
		return context.WithValue(parent, ctxKey{}, v)
	}

	return context.WithValue(parent, ctxKey{}, attrs)
}

// ReplaceAttr is a hook used for modifying attribute values.
//
// Currently it is only replacing passed error with their string form.
func ReplaceAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Value.Kind() == slog.KindAny {
		if err, ok := attr.Value.Any().(error); ok {
			attr.Value = slog.StringValue(err.Error())
		}
	}

	return attr
}
