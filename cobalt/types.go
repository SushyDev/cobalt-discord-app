package cobalt

// RequestOptions represents options for cobalt API requests
type RequestOptions struct {
	URL          string
	VideoQuality string
	AudioFormat  string
	AudioBitrate string
	FilenameStyle string
	DownloadMode string
	YoutubeVideoCodec string
	YoutubeDubLang string
	AlwaysProxy bool
	DisableMetadata bool
	TiktokFullAudio bool
	TiktokH265 bool
	TwitterGif bool
	YoutubeHLS bool
}

// PickerItem represents a single media item in a picker response
type PickerItem struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Thumb string `json:"thumb,omitempty"`
}

// ErrorContext provides additional context for error responses
type ErrorContext struct {
	Service string `json:"service,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// ErrorResponse represents an error from the cobalt API
type ErrorResponse struct {
	Code    string       `json:"code"`
	Context *ErrorContext `json:"context,omitempty"`
}

// CobaltResponse represents a response from the cobalt API
type CobaltResponse struct {
	Status        string       `json:"status"`
	URL           string       `json:"url,omitempty"`
	Filename      string       `json:"filename,omitempty"`
	Audio         string       `json:"audio,omitempty"`
	AudioFilename string       `json:"audioFilename,omitempty"`
	Picker        []PickerItem `json:"picker,omitempty"`
	Error         ErrorResponse `json:"error,omitempty"`
}