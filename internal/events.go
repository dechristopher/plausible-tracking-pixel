package internal

type plausibleEvent struct {
	UserAgent     string `json:"-"`
	XForwardedFor string `json:"-"`
	EventName     string `json:"name,omitempty"`
	URL           string `json:"url,omitempty"`
	Domain        string `json:"domain,omitempty"`
}
