package httpapi

import "fmt"

// chatRequest is the subset of the OpenAI-style chat completions payload the adapter
// validates at its boundary before forwarding to cofiswarm-dispatch.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// validRoles is the set of message roles the adapter accepts.
var validRoles = map[string]bool{
	"system":    true,
	"user":      true,
	"assistant": true,
	"tool":      true,
}

// Validate enforces the required shape of a chat request. It returns a descriptive
// error naming the offending field so callers can surface a 400 with a useful message.
func (c chatRequest) Validate() error {
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(c.Messages) == 0 {
		return fmt.Errorf("messages must contain at least one entry")
	}
	for i, m := range c.Messages {
		if !validRoles[m.Role] {
			return fmt.Errorf("messages[%d].role %q is not one of system, user, assistant, tool", i, m.Role)
		}
		if m.Content == "" {
			return fmt.Errorf("messages[%d].content is required", i)
		}
	}
	return nil
}
