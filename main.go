package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

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

type Engine struct {
	Users       map[string]*User
	Subreddits  map[string]*Subreddit
	Posts       map[int]*Post
	Comments    map[int]*Comment
	VoteHistory map[string]map[int]int
	Mutex       sync.Mutex
	PostID      int
	CommentID   int
}

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

// Add a comment to a post
func (e *Engine) AddComment(postID int, author, content string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Posts[postID]; !exists {
		return "Post does not exist."
	}
	e.CommentID++
	comment := &Comment{
		ID:      e.CommentID,
		Author:  author,
		Content: content,
		Votes:   0,
	}
	e.Posts[postID].Comments = append(e.Posts[postID].Comments, comment)
	e.Comments[e.CommentID] = comment
	return fmt.Sprintf("Comment added successfully with ID %d.", e.CommentID)
}

// Reply to a comment
func (e *Engine) ReplyToComment(commentID int, author, content string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Comments[commentID]; !exists {
		return "Comment does not exist."
	}
	e.CommentID++
	reply := &Comment{
		ID:      e.CommentID,
		Author:  author,
		Content: content,
		Votes:   0,
	}
	e.Comments[commentID].Replies = append(e.Comments[commentID].Replies, reply)
	e.Comments[e.CommentID] = reply
	return fmt.Sprintf("Reply added successfully with ID %d.", e.CommentID)
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
		for _, comment := range post.Comments {
			feed += e.displayComment(comment, 1)
		}
	}
	return feed
}

// Recursive function to display comments and replies
func (e *Engine) displayComment(comment *Comment, depth int) string {
	indent := strings.Repeat("  ", depth)
	result := fmt.Sprintf("%sComment ID: %d | Author: %s | Votes: %d | Content: %s\n", indent, comment.ID, comment.Author, comment.Votes, comment.Content)
	for _, reply := range comment.Replies {
		result += e.displayComment(reply, depth+1)
	}
	return result
}

// Menu for user interaction
func mainMenu(engine *Engine) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- Reddit Clone ---")
		fmt.Println("1. Register User")
		fmt.Println("2. Create Subreddit")
		fmt.Println("3. Create Post")
		fmt.Println("4. Add Comment")
		fmt.Println("5. Reply to Comment")
		fmt.Println("6. View Subreddit Feed")
		fmt.Println("7. Exit")
		fmt.Print("Enter your choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("Enter username: ")
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)
			fmt.Print("Enter password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)
			fmt.Println(engine.RegisterUser(username, password))
		case "2":
			fmt.Print("Enter subreddit name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			fmt.Print("Enter creator username: ")
			creator, _ := reader.ReadString('\n')
			creator = strings.TrimSpace(creator)
			fmt.Println(engine.CreateSubreddit(name, creator))
		case "3":
			fmt.Print("Enter subreddit name: ")
			subreddit, _ := reader.ReadString('\n')
			subreddit = strings.TrimSpace(subreddit)
			fmt.Print("Enter author username: ")
			author, _ := reader.ReadString('\n')
			author = strings.TrimSpace(author)
			fmt.Print("Enter post content: ")
			content, _ := reader.ReadString('\n')
			content = strings.TrimSpace(content)
			fmt.Println(engine.CreatePost(subreddit, author, content))
		case "4":
			fmt.Print("Enter post ID: ")
			postIDStr, _ := reader.ReadString('\n')
			postID, _ := strconv.Atoi(strings.TrimSpace(postIDStr))
			fmt.Print("Enter author username: ")
			author, _ := reader.ReadString('\n')
			author = strings.TrimSpace(author)
			fmt.Print("Enter comment content: ")
			content, _ := reader.ReadString('\n')
			content = strings.TrimSpace(content)
			fmt.Println(engine.AddComment(postID, author, content))
		case "5":
			fmt.Print("Enter comment ID: ")
			commentIDStr, _ := reader.ReadString('\n')
			commentID, _ := strconv.Atoi(strings.TrimSpace(commentIDStr))
			fmt.Print("Enter author username: ")
			author, _ := reader.ReadString('\n')
			author = strings.TrimSpace(author)
			fmt.Print("Enter reply content: ")
			content, _ := reader.ReadString('\n')
			content = strings.TrimSpace(content)
			fmt.Println(engine.ReplyToComment(commentID, author, content))
		case "6":
			fmt.Print("Enter subreddit name: ")
			subreddit, _ := reader.ReadString('\n')
			subreddit = strings.TrimSpace(subreddit)
			fmt.Println(engine.GetFeed(subreddit))
		case "7":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice, please try again.")
		}
	}
}

func main() {
	engine := NewEngine()
	mainMenu(engine)
}
