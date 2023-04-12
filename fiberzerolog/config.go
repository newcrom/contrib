package fiberzerolog

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const (
	FieldReferer       = "referer"
	FieldProtocol      = "protocol"
	FieldPID           = "pid"
	FieldPort          = "port"
	FieldIP            = "ip"
	FieldIPs           = "ips"
	FieldHost          = "host"
	FieldPath          = "path"
	FieldURL           = "url"
	FieldUserAgent     = "ua"
	FieldLatency       = "latency"
	FieldStatus        = "status"
	FieldResBody       = "resBody"
	FieldQueryParams   = "queryParams"
	FieldBody          = "body"
	FieldBytesReceived = "bytesReceived"
	FieldBytesSent     = "bytesSent"
	FieldRoute         = "route"
	FieldMethod        = "method"
	FieldRequestID     = "requestId"
	FieldError         = "error"
	FieldReqHeaders    = "reqHeaders"
)

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *fiber.Ctx) bool

	// SkipBody defines a function to skip log  "body" field when returned true.
	//
	// Optional. Default: nil
	SkipBody func(c *fiber.Ctx) bool

	// SkipResBody defines a function to skip log  "resBody" field when returned true.
	//
	// Optional. Default: nil
	SkipResBody func(c *fiber.Ctx) bool

	// GetResBody defines a function to get ResBody.
	//  eg: when use compress middleware, resBody is unreadable. you can set GetResBody func to get readable resBody.
	//
	// Optional. Default: nil
	GetResBody func(c *fiber.Ctx) []byte

	// Skip logging for these uri
	//
	// Optional. Default: nil
	SkipURIs []string

	// Add custom zerolog logger.
	//
	// Optional. Default: zerolog.New(os.Stderr).With().Timestamp().Logger()
	Logger *zerolog.Logger

	// GetLogger defines a function to get custom zerolog logger.
	//  eg: when we need to create a new logger for each request.
	//
	// GetLogger will override Logger.
	//
	// Optional. Default: nil
	GetLogger func(c *fiber.Ctx) zerolog.Logger

	// Add fields what you want see.
	//
	// Optional. Default: {"latency", "status", "method", "url", "error"}
	Fields []string

	// Custom response messages.
	// Response codes >= 500 will be logged with Messages[0].
	// Response codes >= 400 will be logged with Messages[1].
	// Other response codes will be logged with Messages[2].
	// You can specify less, than 3 messages, but you must specify at least 1.
	// Specifying more than 3 messages is useless.
	//
	// Optional. Default: {"Server error", "Client error", "Success"}
	Messages []string

	// Custom response levels.
	// Response codes >= 500 will be logged with Levels[0].
	// Response codes >= 400 will be logged with Levels[1].
	// Other response codes will be logged with Levels[2].
	// You can specify less, than 3 levels, but you must specify at least 1.
	// Specifying more than 3 levels is useless.
	//
	// Optional. Default: {zerolog.ErrorLevel, zerolog.WarnLevel, zerolog.InfoLevel}
	Levels []zerolog.Level
}

func (c *Config) loggerCtx(fc *fiber.Ctx) zerolog.Context {
	if c.GetLogger != nil {
		return c.GetLogger(fc).With()
	}

	return c.Logger.With()
}

func (c *Config) logger(fc *fiber.Ctx, latency time.Duration, err error) zerolog.Logger {
	zc := c.loggerCtx(fc)

	for _, field := range c.Fields {
		switch field {
		case FieldReferer:
			zc = zc.Str(FieldReferer, fc.Get(fiber.HeaderReferer))
		case FieldProtocol:
			zc = zc.Str(FieldProtocol, fc.Protocol())
		case FieldPID:
			zc = zc.Int(FieldPID, os.Getpid())
		case FieldPort:
			zc = zc.Str(FieldPort, fc.Port())
		case FieldIP:
			zc = zc.Str(FieldIP, fc.IP())
		case FieldIPs:
			zc = zc.Str(FieldIPs, fc.Get(fiber.HeaderXForwardedFor))
		case FieldHost:
			zc = zc.Str(FieldHost, fc.Hostname())
		case FieldPath:
			zc = zc.Str(FieldPath, fc.Path())
		case FieldURL:
			zc = zc.Str(FieldURL, fc.OriginalURL())
		case FieldUserAgent:
			zc = zc.Str(FieldUserAgent, fc.Get(fiber.HeaderUserAgent))
		case FieldLatency:
			zc = zc.Dur(FieldLatency, latency)
		case FieldStatus:
			zc = zc.Int(FieldStatus, fc.Response().StatusCode())
		case FieldResBody:
			if c.SkipResBody == nil || !c.SkipResBody(fc) {
				if c.GetResBody == nil {
					zc = zc.Bytes(FieldResBody, fc.Response().Body())
				} else {
					zc = zc.Bytes(FieldResBody, c.GetResBody(fc))
				}
			}
		case FieldQueryParams:
			zc = zc.Stringer(FieldQueryParams, fc.Request().URI().QueryArgs())
		case FieldBody:
			if c.SkipBody == nil || !c.SkipBody(fc) {
				zc = zc.Bytes(FieldBody, fc.Body())
			}
		case FieldBytesReceived:
			zc = zc.Int(FieldBytesReceived, len(fc.Request().Body()))
		case FieldBytesSent:
			zc = zc.Int(FieldBytesSent, len(fc.Response().Body()))
		case FieldRoute:
			zc = zc.Str(FieldRoute, fc.Route().Path)
		case FieldMethod:
			zc = zc.Str(FieldMethod, fc.Method())
		case FieldRequestID:
			zc = zc.Str(FieldRequestID, fc.GetRespHeader(fiber.HeaderXRequestID))
		case FieldError:
			if err != nil {
				zc = zc.Err(err)
			}
		case FieldReqHeaders:
			fc.Request().Header.VisitAll(func(k, v []byte) {
				zc = zc.Bytes(string(k), v)
			})
		}
	}

	return zc.Logger()
}

var logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next:     nil,
	Logger:   &logger,
	Fields:   []string{FieldLatency, FieldStatus, FieldMethod, FieldURL, FieldError},
	Messages: []string{"Server error", "Client error", "Success"},
	Levels:   []zerolog.Level{zerolog.ErrorLevel, zerolog.WarnLevel, zerolog.InfoLevel},
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Next == nil {
		cfg.Next = ConfigDefault.Next
	}

	if cfg.Logger == nil {
		cfg.Logger = ConfigDefault.Logger
	}

	if cfg.Fields == nil {
		cfg.Fields = ConfigDefault.Fields
	}

	if cfg.Messages == nil {
		cfg.Messages = ConfigDefault.Messages
	}

	if cfg.Levels == nil {
		cfg.Levels = ConfigDefault.Levels
	}

	return cfg
}
