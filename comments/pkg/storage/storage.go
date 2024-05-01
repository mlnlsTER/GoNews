package comStorage

type Comment struct {
    ID      int    // comment number
    ID_News   int64 // news number
    ID_Parent int64 // parent number (if the answer to the comment) 
    Content string  // comment content
    ComTime    int64 // comment time
}

// Interface specifies the contract for working with the database.
type CommentsInterface interface {
	Comments(int64) ([]Comment, error) // Get comments comments from the database.
	AddComments([]Comment) error     // Add comments to the database.
}
