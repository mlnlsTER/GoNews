package newsStorage

// Publication retrieved from RSS.
type Post struct {
	ID      int    // record number
	Title   string // publication title
	Content string // publication content
	PubTime int64  // publication time
	Link    string // publication link
}

type Pagination struct {
    TotalPages   int
    CurrentPage  int
    PageSize     int
}

// NewsInterface specifies the contract for working with the database.
type NewsInterface interface {
	Posts(int, string) ([]Post, Pagination, error) // Get publications from the database.
	AddPosts([]Post) error     // Add publications to the database.
	PostDetail(int) (*Post, error) // Get detailed publication
}
