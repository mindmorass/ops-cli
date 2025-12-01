package startpage

// StartpageConfig represents the startpage configuration
type StartpageConfig struct {
	Name      string          `json:"name"`
	Path      string          `json:"path"`
	Theme     string          `json:"theme"`
	Bookmarks []BookmarkGroup `json:"bookmarks"`
	Settings  Settings        `json:"settings"`
}

// BookmarkGroup represents a group of bookmarks
type BookmarkGroup struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Icon      string     `json:"icon,omitempty"`
	Color     string     `json:"color,omitempty"`
	Bookmarks []Bookmark `json:"bookmarks"`
}

// Bookmark represents a single bookmark
type Bookmark struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// Settings represents startpage settings
type Settings struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Favicon     string `json:"favicon,omitempty"`
	Background  string `json:"background,omitempty"`
	Author      Author `json:"author"`
}

// Author represents the author information
type Author struct {
	Name  string `json:"name"`
	Photo string `json:"photo,omitempty"`
	Bio   string `json:"bio,omitempty"`
}

// OpenLinksConfig represents the OpenLinks.json format
type OpenLinksConfig struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	URLBase     string     `json:"url_base"`
	Theme       string     `json:"theme"`
	Footer      string     `json:"footer"`
	Profile     Profile    `json:"profile"`
	Links       []OpenLink `json:"links"`
}

// Profile represents the OpenLinks profile
type Profile struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Adult       bool   `json:"adult"`
}

// OpenLink represents a link in OpenLinks format
type OpenLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon"`
}
