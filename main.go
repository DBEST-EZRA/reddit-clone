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
	Username     string
	Password     string
	Karma        int
	DirectMessages []Message
}

type Message struct {
	Sender  string
	Content string
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
	Users          map[string]*User
	Subreddits     map[string]*Subreddit
	Posts          map[int]*Post
	Comments       map[int]*Comment
	VoteHistory    map[string]map[int]int
	Mutex          sync.Mutex
	PostID         int
	CommentID      int
	DisconnectedUsers map[string]bool
}

func NewEngine() *Engine {
	return &Engine{
		Users:          make(map[string]*User),
		Subreddits:     make(map[string]*Subreddit),
		Posts:          make(map[int]*Post),
		Comments:       make(map[int]*Comment),
		VoteHistory:    make(map[string]map[int]int),
		DisconnectedUsers: make(map[string]bool),
	}
}

//Added Code lies here

//START
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
//STOP

// --- ADDITIONAL FUNCTIONALITY STARTS HERE ---

// Upvote or Downvote a Post
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
		e.Users[e.Posts[postID].Author].Karma -= previousVote
	}
	e.Posts[postID].Votes += vote
	e.VoteHistory[username][postID] = vote
	e.Users[e.Posts[postID].Author].Karma += vote
	return "Vote registered successfully."
}

// Send a Direct Message
func (e *Engine) SendMessage(sender, recipient, content string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Users[recipient]; !exists {
		return "Recipient does not exist."
	}
	message := Message{Sender: sender, Content: content}
	e.Users[recipient].DirectMessages = append(e.Users[recipient].DirectMessages, message)
	return "Message sent successfully."
}

// List Direct Messages
func (e *Engine) ListMessages(username string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Users[username]; !exists {
		return "User does not exist."
	}
	messages := e.Users[username].DirectMessages
	if len(messages) == 0 {
		return "No messages."
	}
	result := fmt.Sprintf("Direct messages for %s:\n", username)
	for i, msg := range messages {
		result += fmt.Sprintf("%d. From %s: %s\n", i+1, msg.Sender, msg.Content)
	}
	return result
}

// Simulate User Connection/Disconnection
func (e *Engine) SimulateConnection(username string, connected bool) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Users[username]; !exists {
		return "User does not exist."
	}
	e.DisconnectedUsers[username] = !connected
	if connected {
		return fmt.Sprintf("%s is now connected.", username)
	}
	return fmt.Sprintf("%s is now disconnected.", username)
}

// Simulate Zipf Distribution for Subreddit Membership
func (e *Engine) SimulateZipfDistribution() string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	for i := 1; i <= 10; i++ {
		subredditName := fmt.Sprintf("subreddit_%d", i)
		if _, exists := e.Subreddits[subredditName]; !exists {
			e.Subreddits[subredditName] = &Subreddit{
				Name:    subredditName,
				Members: make(map[string]bool),
				Posts:   []*Post{},
			}
		}
		members := 100 / i // Zipf-like distribution
		for j := 1; j <= members; j++ {
			username := fmt.Sprintf("user_%d", j)
			e.Subreddits[subredditName].Members[username] = true
		}
	}
	return "Simulated Zipf distribution for subreddit membership."
}

// --- MENU INTEGRATION ---

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
		fmt.Println("7. Upvote/Downvote Post")
		fmt.Println("8. Send Direct Message")
		fmt.Println("9. List Direct Messages")
		fmt.Println("10. Simulate Connection/Disconnection")
		fmt.Println("11. Simulate Zipf Distribution")
		fmt.Println("12. Exit")
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
			fmt.Print("Enter post ID: ")
			postIDStr, _ := reader.ReadString('\n')
			postID, _ := strconv.Atoi(strings.TrimSpace(postIDStr))
			fmt.Print("Enter username: ")
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)
			fmt.Print("Enter vote (1 for upvote, -1 for downvote): ")
			voteStr, _ := reader.ReadString('\n')
			vote, _ := strconv.Atoi(strings.TrimSpace(voteStr))
			fmt.Println(engine.VotePost(username, postID, vote))
		case "8":
			fmt.Print("Enter sender username: ")
			sender, _ := reader.ReadString('\n')
			sender = strings.TrimSpace(sender)
			fmt.Print("Enter recipient username: ")
			recipient, _ := reader.ReadString('\n')
			recipient = strings.TrimSpace(recipient)
			fmt.Print("Enter message content: ")
			content, _ := reader.ReadString('\n')
			content = strings.TrimSpace(content)
			fmt.Println(engine.SendMessage(sender, recipient, content))
		case "9":
			fmt.Print("Enter username: ")
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)
			fmt.Println(engine.ListMessages(username))
		case "10":
			fmt.Print("Enter username: ")
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)
			fmt.Print("Enter connection status (true for connect, false for disconnect): ")
			statusStr, _ := reader.ReadString('\n')
			connected, _ := strconv.ParseBool(strings.TrimSpace(statusStr))
			fmt.Println(engine.SimulateConnection(username, connected))
		case "11":
			fmt.Println(engine.SimulateZipfDistribution())
		case "12":
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
