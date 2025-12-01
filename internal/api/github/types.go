package github

// RepositoryListOptions represents options for listing repositories
type RepositoryListOptions struct {
	User      string
	Org       string
	Query     string
	Type      string
	Sort      string
	Direction string
	PerPage   int
	Page      int
}

// FileSearchOptions represents options for searching files
type FileSearchOptions struct {
	Owner     string
	Repo      string
	Query     string
	Path      string
	Filename  string
	Extension string
	PerPage   int
	Page      int
}
