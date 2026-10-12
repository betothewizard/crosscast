package broadcast

// AuthorPayload represents author metadata in a rich embed.
type AuthorPayload struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// FieldPayload represents a key-value field in a rich embed.
type FieldPayload struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// FooterPayload represents footer metadata in a rich embed.
type FooterPayload struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// BroadcastPayload represents an inbound broadcast request.
type BroadcastPayload struct {
	Platform    string         `json:"platform,omitempty"`    // "all" (default), "discord", or "telegram"
	Target      string         `json:"target"`                // Target alias ("jobs", "alerts") or direct channel ID
	Content     string         `json:"content,omitempty"`     // Raw content or mentions (e.g. "@everyone")
	Title       string         `json:"title,omitempty"`       // Embed Title
	Description string         `json:"description,omitempty"` // Embed Description (Markdown)
	URL         string         `json:"url,omitempty"`         // Embed Title URL
	Color       int            `json:"color,omitempty"`       // Hex integer (e.g. 0x10B981)
	Author      *AuthorPayload `json:"author,omitempty"`
	Thumbnail   string         `json:"thumbnail,omitempty"`   // URL of thumbnail image
	Image       string         `json:"image,omitempty"`       // URL of banner image
	Fields      []FieldPayload `json:"fields,omitempty"`
	Footer      *FooterPayload `json:"footer,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`   // Optional ISO8601 string
}
