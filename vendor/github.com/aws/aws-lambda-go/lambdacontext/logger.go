//go:build go1.21
// +build go1.21

// Copyright 2026 Amazon.com, Inc. or its affiliates. All Rights Reserved.

package lambdacontext

import (
	"context"
	"log/slog"
	"os"
)

// logFormat is the log format from AWS_LAMBDA_LOG_FORMAT (TEXT or JSON)
var logFormat = os.Getenv("AWS_LAMBDA_LOG_FORMAT")

// logLevel is the log level from AWS_LAMBDA_LOG_LEVEL
var logLevel = os.Getenv("AWS_LAMBDA_LOG_LEVEL")

// field represents a Lambda context field to include in log records.
type field struct {
	key   string
	value func(*LambdaContext) string
}

// logOptions holds configuration for the Lambda log handler.
type logOptions struct {
	fields []field
}

// LogOption is a functional option for configuring the Lambda log handler.
type LogOption func(*logOptions)

// WithFunctionARN includes the invoked function ARN in log records.
func WithFunctionARN() LogOption {
	return func(o *logOptions) {
		o.fields = append(o.fields, field{"functionArn", func(lc *LambdaContext) string { return lc.InvokedFunctionArn }})
	}
}

// WithTenantID includes the tenant ID in log records (for multi-tenant functions).
func WithTenantID() LogOption {
	return func(o *logOptions) {
		o.fields = append(o.fields, field{"tenantId", func(lc *LambdaContext) string { return lc.TenantID }})
	}
}

// NewLogHandler returns a [slog.Handler] for AWS Lambda structured logging.
// It reads AWS_LAMBDA_LOG_FORMAT and AWS_LAMBDA_LOG_LEVEL from environment,
// and injects requestId from Lambda context into each log record.
//
// By default, only requestId is injected. Use WithFunctionARN or WithTenantID to include more.
// See the package examples for usage.
func NewLogHandler(opts ...LogOption) slog.Handler {
	options := &logOptions{}
	for _, opt := range opts {
		opt(options)
	}

	level := parseLogLevel()
	handlerOpts := &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: ReplaceAttr,
	}

	var h slog.Handler
	if logFormat == "JSON" {
		h = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		h = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	return &lambdaHandler{handler: h, fields: options.fields}
}

// NewLogger returns a [*slog.Logger] configured for AWS Lambda structured logging.
// This is a convenience function equivalent to slog.New(NewLogHandler(opts...)).
func NewLogger(opts ...LogOption) *slog.Logger {
	return slog.New(NewLogHandler(opts...))
}

// ReplaceAttr maps slog's default keys to AWS Lambda's log format (time->timestamp, msg->message).
func ReplaceAttr(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return attr
	}

	switch attr.Key {
	case slog.TimeKey:
		attr.Key = "timestamp"
	case slog.MessageKey:
		attr.Key = "message"
	}
	return attr
}

// lambdaHandler wraps a slog.Handler to inject Lambda context fields.
type lambdaHandler struct {
	handler slog.Handler
	fields  []field
}

// Enabled implements slog.Handler.
func (h *lambdaHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler.
func (h *lambdaHandler) Handle(ctx context.Context, r slog.Record) error {
	if lc, ok := FromContext(ctx); ok {
		r.AddAttrs(slog.String("requestId", lc.AwsRequestID))

		for _, field := range h.fields {
			if v := field.value(lc); v != "" {
				r.AddAttrs(slog.String(field.key, v))
			}
		}
	}
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements slog.Handler.
func (h *lambdaHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &lambdaHandler{
		handler: h.handler.WithAttrs(attrs),
		fields:  h.fields,
	}
}

// WithGroup implements slog.Handler.
func (h *lambdaHandler) WithGroup(name string) slog.Handler {
	return &lambdaHandler{
		handler: h.handler.WithGroup(name),
		fields:  h.fields,
	}
}

func parseLogLevel() slog.Level {
	switch logLevel {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
