package claude

import (
	"context"
	"errors"

	"github.com/anthropics/anthropic-sdk-go"
)

// Every call to Claude goes through send, so a meter can count calls and
// tokens and hold the day to a cap the learner sets.

// Meter counts Claude's use: Allow says whether another call fits today's
// cap, Used records a call's tokens.
type Meter interface {
	Allow() bool
	Used(model string, input, output int64)
}

// ErrDailyCap is returned once today's calls have reached the cap set in
// settings.
var ErrDailyCap = errors.New("today's Claude limit is reached (change it in settings, the , key at camp)")

// send makes one call to Claude, within the meter's cap, and records it.
func (c *Client) send(ctx context.Context, params anthropic.MessageNewParams) (*anthropic.Message, error) {
	if c.Meter != nil && !c.Meter.Allow() {
		return nil, ErrDailyCap
	}
	msg, err := c.client.Messages.New(ctx, params)
	if err != nil {
		if c.server {
			err = serverError(err)
		}
		return nil, err
	}
	if c.Meter != nil {
		c.Meter.Used(string(params.Model), msg.Usage.InputTokens, msg.Usage.OutputTokens)
	}
	return msg, nil
}
