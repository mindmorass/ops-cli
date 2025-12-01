package confluence

// Page represents a Confluence page
type Page struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Status    string   `json:"status"`
	Title     string   `json:"title"`
	Body      *Content `json:"body,omitempty"`
	Version   *Version `json:"version,omitempty"`
	Space     *Space   `json:"space,omitempty"`
	Ancestors []Page   `json:"ancestors,omitempty"`
	Links     *Links   `json:"_links,omitempty"`
}

// Content represents page content in different representations
type Content struct {
	Storage    *ContentValue `json:"storage,omitempty"`
	View       *ContentValue `json:"view,omitempty"`
	ExportView *ContentValue `json:"export_view,omitempty"`
}

// ContentValue represents a content value
type ContentValue struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// Version represents page version information
type Version struct {
	By     *User  `json:"by,omitempty"`
	When   string `json:"when"`
	Number int    `json:"number"`
}

// Space represents a Confluence space
type Space struct {
	ID   int    `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// User represents a Confluence user
type User struct {
	Type        string `json:"type"`
	AccountID   string `json:"accountId,omitempty"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

// Comment represents a Confluence comment
type Comment struct {
	ID      string   `json:"id"`
	Title   string   `json:"title,omitempty"`
	Body    *Content `json:"body,omitempty"`
	Version *Version `json:"version,omitempty"`
}

// Attachment represents a Confluence attachment
type Attachment struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// Metadata represents attachment metadata
type Metadata struct {
	FileSize int64 `json:"fileSize,omitempty"`
}

// Links represents API links
type Links struct {
	WebUI string `json:"webui,omitempty"`
}

// SpacesResponse represents the response from getting spaces
type SpacesResponse struct {
	Results []Space `json:"results"`
	Size    int     `json:"size"`
}

// PagesResponse represents the response from getting pages
type PagesResponse struct {
	Results []Page `json:"results"`
	Size    int    `json:"size"`
}

// CommentsResponse represents the response from getting comments
type CommentsResponse struct {
	Results []Comment `json:"results"`
	Size    int       `json:"size"`
}

// AttachmentsResponse represents the response from getting attachments
type AttachmentsResponse struct {
	Results []Attachment `json:"results"`
	Size    int          `json:"size"`
}
