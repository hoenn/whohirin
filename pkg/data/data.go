package data

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/hoenn/go-hn/pkg/hnapi"
)

// Fetcher gathers and caches story and comment data for a user. Fetcher lazy loads data into the
// underlying data structures as it is requested.
type Fetcher struct {
	hn     *hnapi.HNClient
	userID string
	posts  map[string]*Post
}

type Post struct {
	ID       string
	Title    string
	comments map[string]*Comment
}

type Comment struct {
	Data *hnapi.Comment
	Read bool
}

// NewFetcher constructs and initializes a fetcher with the given options.
func NewFetcher(userID string) (*Fetcher, error) {
	client := hnapi.NewHNClient()
	user, err := client.User(userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("unable to fetch user submissions: %w", err)
	}

	// Initialize convert IDs to strings and initialize posts with nil
	posts := make(map[string]*Post)
	for _, s := range user.Submitted {
		postIDStr := fmt.Sprintf("%d", s)
		posts[postIDStr] = nil
	}

	return &Fetcher{
		hn:     client,
		userID: userID,
		posts:  posts,
	}, nil
}

// PostList returns keys in descending order.
func (f *Fetcher) PostList() []string {
	keys := make([]string, 0, len(f.posts))
	for p := range f.posts {
		keys = append(keys, p)
	}
	sort.Slice(keys, func(i, j int) bool {
		// these keys are returned from the API. A lot of things would break
		// if they weren't integers.
		a, _ := strconv.Atoi(keys[i])
		b, _ := strconv.Atoi(keys[j])
		return a > b
	})
	return keys
}

// Post will return a Post preloaded with title for the given ID. If it has been already
// requested then a cached Post will be returned. If not, it will be fetched using the client.
func (f *Fetcher) Post(id string) (*Post, error) {
	post, found := f.posts[id]
	if !found {
		return nil, fmt.Errorf("not found")
	}
	if post != nil {
		return post, nil
	}

	item, err := f.hn.Item(id)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch post: %w", err)
	}
	story, ok := item.(*hnapi.Story)
	if !ok {
		return nil, fmt.Errorf("id was not a post")
	}
	p := &Post{
		Title: story.Title,
		ID:    id,
	}

	comments := make(map[string]*Comment)
	for _, k := range story.Kids {
		commIDStr := fmt.Sprintf("%d", k)
		comments[commIDStr] = nil
	}
	p.comments = comments
	return p, nil
}

// PostCommentsList will lazy load a Post into the cache and return a list of comment
// IDs for that post in descending order.
func (f *Fetcher) PostCommentsList(postID string) ([]string, error) {
	p, err := f.Post(postID)
	if err != nil {
		return []string{}, fmt.Errorf("could not find post: %w", err)
	}

	keys := make([]string, 0, len(p.comments))
	for c := range p.comments {
		keys = append(keys, c)
	}
	sort.Slice(keys, func(i, j int) bool {
		// these keys are returned from the API. A lot of things would break
		// if they weren't integers.
		a, _ := strconv.Atoi(keys[i])
		b, _ := strconv.Atoi(keys[j])
		return a > b
	})
	return keys, nil
}

func (f *Fetcher) PostComment(postID, commentID string) (*Comment, error) {
	p, err := f.Post(postID)
	if err != nil {
		return nil, fmt.Errorf("could not find post: %w", err)
	}
	c, found := p.comments[commentID]
	if !found {
		return nil, fmt.Errorf("not found")
	}
	if c != nil {
		return c, nil
	}

	item, err := f.hn.Item(commentID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch comment: %w", err)
	}
	comment, ok := item.(*hnapi.Comment)
	if !ok {
		return nil, fmt.Errorf("id was not a comment")
	}

	cc := &Comment{
		Data: comment,
		Read: false,
	}
	p.comments[commentID] = cc

	return cc, nil
}
