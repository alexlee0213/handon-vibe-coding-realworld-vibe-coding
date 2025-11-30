package domain

// Tag represents a tag that can be associated with articles
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TagsResponse represents the tags list returned to clients (RealWorld API format)
type TagsResponse struct {
	Tags []string `json:"tags"`
}

// NewTagsResponse creates a new TagsResponse from a list of tag names
func NewTagsResponse(tagNames []string) *TagsResponse {
	if tagNames == nil {
		tagNames = []string{}
	}
	return &TagsResponse{
		Tags: tagNames,
	}
}
