package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Structs for core entities
type User struct {
	Username string
	Password string
	Karma    int
}

type Subreddit struct {
	Name    string
	Members map[string]bool
	Posts   []*Post
}

type Post struct {
	ID       int
	Author   string
	Content  string
	Votes    int
	Comments []*Comment
}

type Comment struct {
	ID       int
	Author   string
	Content  string
	Votes    int
	Replies  []*Comment
}

// Engine struct to store data
type Engine struct {
	Users       map[string]*User
	Subreddits  map[string]*Subreddit
	Posts       map[int]*Post
	Comments    map[int]*Comment
	VoteHistory map[string]map[int]int // Track user votes on posts and comments
	Mutex       sync.Mutex
	PostID      int
	CommentID   int
}

// Initialize the Engine
func NewEngine() *Engine {
	return &Engine{
		Users:       make(map[string]*User),
		Subreddits:  make(map[string]*Subreddit),
		Posts:       make(map[int]*Post),
		Comments:    make(map[int]*Comment),
		VoteHistory: make(map[string]map[int]int),
	}
}

// Register a user
func (e *Engine) RegisterUser(username, password string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Users[username]; exists {
		return "Username already exists."
	}
	e.Users[username] = &User{Username: username, Password: password, Karma: 0}
	return "User registered successfully."
}

// Create a subreddit
func (e *Engine) CreateSubreddit(name, creator string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Subreddits[name]; exists {
		return "Subreddit already exists."
	}
	e.Subreddits[name] = &Subreddit{
		Name:    name,
		Members: map[string]bool{creator: true},
		Posts:   []*Post{},
	}
	return "Subreddit created successfully."
}

// Post to a subreddit
func (e *Engine) CreatePost(subreddit, author, content string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Subreddits[subreddit]; !exists {
		return "Subreddit does not exist."
	}
	e.PostID++
	post := &Post{
		ID:      e.PostID,
		Author:  author,
		Content: content,
		Votes:   0,
	}
	e.Subreddits[subreddit].Posts = append(e.Subreddits[subreddit].Posts, post)
	e.Posts[e.PostID] = post
	return fmt.Sprintf("Post created successfully with ID %d.", e.PostID)
}

// Simulate voting
func (e *Engine) VotePost(username string, postID, vote int) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Posts[postID]; !exists {
		return "Post does not exist."
	}
	if e.VoteHistory[username] == nil {
		e.VoteHistory[username] = make(map[int]int)
	}
	if previousVote, voted := e.VoteHistory[username][postID]; voted {
		e.Posts[postID].Votes -= previousVote
	}
	e.Posts[postID].Votes += vote
	e.VoteHistory[username][postID] = vote
	return "Vote registered successfully."
}

// Display a subreddit feed
func (e *Engine) GetFeed(subreddit string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Subreddits[subreddit]; !exists {
		return "Subreddit does not exist."
	}
	feed := fmt.Sprintf("Feed for Subreddit: %s\n", subreddit)
	for _, post := range e.Subreddits[subreddit].Posts {
		feed += fmt.Sprintf("Post ID: %d | Author: %s | Votes: %d | Content: %s\n", post.ID, post.Author, post.Votes, post.Content)
	}
	return feed
}

// Simulate user behavior
func SimulateUser(engine *Engine, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	username := fmt.Sprintf("user_%d", id)
	password := "password"
	engine.RegisterUser(username, password)

	// Join or create subreddits
	for i := 0; i < rand.Intn(5)+1; i++ {
		subreddit := fmt.Sprintf("subreddit_%d", rand.Intn(10)+1)
		engine.CreateSubreddit(subreddit, username)
	}

	// Post content
	for i := 0; i < rand.Intn(5)+1; i++ {
		subreddit := fmt.Sprintf("subreddit_%d", rand.Intn(10)+1)
		content := fmt.Sprintf("Post from %s", username)
		engine.CreatePost(subreddit, username, content)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
	}

	// Vote on posts
	for i := 0; i < rand.Intn(10)+1; i++ {
		postID := rand.Intn(50) + 1
		vote := rand.Intn(2)*2 - 1 // Either -1 or +1
		engine.VotePost(username, postID, vote)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	engine := NewEngine()

	var wg sync.WaitGroup
	numUsers := 100

	// Simulate multiple users
	for i := 1; i <= numUsers; i++ {
		wg.Add(1)
		go SimulateUser(engine, i, &wg)
	}

	wg.Wait()

	// Display results for a specific subreddit
	fmt.Println(engine.GetFeed("subreddit_1"))
}
