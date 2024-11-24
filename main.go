package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type User struct {
	Username      string
	Password      string
	Karma         int
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
	Users             map[string]*User
	Subreddits        map[string]*Subreddit
	Posts             map[int]*Post
	Comments          map[int]*Comment
	VoteHistory       map[string]map[int]int
	Mutex             sync.Mutex
	PostID            int
	CommentID         int
	DisconnectedUsers map[string]bool
}

func NewEngine() *Engine {
	return &Engine{
		Users:             make(map[string]*User),
		Subreddits:        make(map[string]*Subreddit),
		Posts:             make(map[int]*Post),
		Comments:          make(map[int]*Comment),
		VoteHistory:       make(map[string]map[int]int),
		DisconnectedUsers: make(map[string]bool),
	}
}

//enabling Cross Site Origin
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Frontend origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true") 
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// Register User API Endpoint
func (e *Engine) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	response := e.RegisterUser(username, password)
	writeResponse(w, response)
}

func (e *Engine) RegisterUser(username, password string) string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	if _, exists := e.Users[username]; exists {
		return "Username already exists."
	}
	e.Users[username] = &User{Username: username, Password: password, Karma: 0}
	return "User registered successfully."
}

// Create Subreddit API Endpoint
func (e *Engine) CreateSubredditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	creator := r.URL.Query().Get("creator")
	if name == "" || creator == "" {
		http.Error(w, "Subreddit name and creator are required", http.StatusBadRequest)
		return
	}
	response := e.CreateSubreddit(name, creator)
	writeResponse(w, response)
}

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

// Create Post API Endpoint
func (e *Engine) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	subreddit := r.URL.Query().Get("subreddit")
	author := r.URL.Query().Get("author")
	content := r.URL.Query().Get("content")
	if subreddit == "" || author == "" || content == "" {
		http.Error(w, "Subreddit, author, and content are required", http.StatusBadRequest)
		return
	}
	response := e.CreatePost(subreddit, author, content)
	writeResponse(w, response)
}

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

// Add Comment API Endpoint
func (e *Engine) AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	postIDStr := r.URL.Query().Get("postID")
	author := r.URL.Query().Get("author")
	content := r.URL.Query().Get("content")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	if author == "" || content == "" {
		http.Error(w, "Author and content are required", http.StatusBadRequest)
		return
	}
	response := e.AddComment(postID, author, content)
	writeResponse(w, response)
}

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

// Reply to Comment API Endpoint
func (e *Engine) ReplyToCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	commentIDStr := r.URL.Query().Get("commentID")
	author := r.URL.Query().Get("author")
	content := r.URL.Query().Get("content")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	if author == "" || content == "" {
		http.Error(w, "Author and content are required", http.StatusBadRequest)
		return
	}
	response := e.ReplyToComment(commentID, author, content)
	writeResponse(w, response)
}

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

// Printing out subreddit feed
func (e *Engine) GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	subreddit := r.URL.Query().Get("subreddit")
	if subreddit == "" {
		http.Error(w, "Subreddit name is required", http.StatusBadRequest)
		return
	}
	response := e.GetFeed(subreddit)
	writeResponse(w, response)
}

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

// Displaying comments and replies
func (e *Engine) displayComment(comment *Comment, depth int) string {
	indent := strings.Repeat("  ", depth)
	result := fmt.Sprintf("%sComment ID: %d | Author: %s | Votes: %d | Content: %s\n", indent, comment.ID, comment.Author, comment.Votes, comment.Content)
	for _, reply := range comment.Replies {
		result += e.displayComment(reply, depth+1)
	}
	return result
}

// Upvote or Downvote 
func (e *Engine) VotePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	postIDStr := r.URL.Query().Get("postID")
	voteStr := r.URL.Query().Get("vote")
	postID, err := strconv.Atoi(postIDStr)
	vote, err2 := strconv.Atoi(voteStr)
	if err != nil || err2 != nil {
		http.Error(w, "Invalid post ID or vote", http.StatusBadRequest)
		return
	}
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	response := e.VotePost(username, postID, vote)
	writeResponse(w, response)
}

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

// Sending Direct Message
func (e *Engine) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	sender := r.URL.Query().Get("sender")
	recipient := r.URL.Query().Get("recipient")
	content := r.URL.Query().Get("content")
	if sender == "" || recipient == "" || content == "" {
		http.Error(w, "Sender, recipient, and content are required", http.StatusBadRequest)
		return
	}
	response := e.SendMessage(sender, recipient, content)
	writeResponse(w, response)
}

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

// Listing all Direct Messages
func (e *Engine) ListMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	response := e.ListMessages(username)
	writeResponse(w, response)
}

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

// Simulating a user Connection or Disconnection
func (e *Engine) SimulateConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")
	connectedStr := r.URL.Query().Get("connected")
	connected, err := strconv.ParseBool(connectedStr)
	if err != nil || username == "" {
		http.Error(w, "Invalid connection status or username", http.StatusBadRequest)
		return
	}
	response := e.SimulateConnection(username, connected)
	writeResponse(w, response)
}

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

// Simulating Zipf Distribution for Subreddit Membership function
func (e *Engine) SimulateZipfDistributionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	response := e.SimulateZipfDistribution()
	writeResponse(w, response)
}

func (e *Engine) SimulateZipfDistribution() string {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	postID := e.PostID
	for i := 1; i <= 10; i++ {
		subredditName := fmt.Sprintf("subreddit_%d", i)
		if _, exists := e.Subreddits[subredditName]; !exists {
			e.Subreddits[subredditName] = &Subreddit{
				Name:    subredditName,
				Members: make(map[string]bool),
				Posts:   []*Post{},
			}
		}
		// Simulating Zipf distribution for membership
		members := 100 / i // more members for higher rank
		for j := 1; j <= members; j++ {
			username := fmt.Sprintf("user_%d", j)
			if _, exists := e.Users[username]; !exists {
				e.Users[username] = &User{Username: username, Karma: 0}
			}
			e.Subreddits[subredditName].Members[username] = true
		}
		// Increasing post frequency for the popular subreddits
		postCount := members / 10
		for p := 0; p < postCount; p++ {
			author := fmt.Sprintf("user_%d", rand.Intn(members)+1)
			content := fmt.Sprintf("Post %d in %s by %s", postID, subredditName, author)
			post := &Post{
				ID:       postID,
				Author:   author,
				Content:  content,
				Votes:    rand.Intn(100), 
				Comments: []*Comment{},
			}
			e.Subreddits[subredditName].Posts = append(e.Subreddits[subredditName].Posts, post)
			e.Posts[postID] = post
			postID++
		}
	}

	// Reposting
	for _, subreddit := range e.Subreddits {
		if len(subreddit.Posts) > 0 {
			repostCount := len(subreddit.Posts) / 5 
			for r := 0; r < repostCount; r++ {
				// Randomly selecting a post from another subreddit
				sourceSubredditName := fmt.Sprintf("subreddit_%d", rand.Intn(10)+1)
				if sourceSubredditName != subreddit.Name {
					sourceSubreddit := e.Subreddits[sourceSubredditName]
					if len(sourceSubreddit.Posts) > 0 {
						randomPost := sourceSubreddit.Posts[rand.Intn(len(sourceSubreddit.Posts))]
						repost := &Post{
							ID:       postID,
							Author:   randomPost.Author,
							Content:  "[Repost] " + randomPost.Content,
							Votes:    randomPost.Votes / 2, 
							Comments: []*Comment{},
						}
						subreddit.Posts = append(subreddit.Posts, repost)
						e.Posts[postID] = repost
						postID++
					}
				}
			}
		}
	}
	e.PostID = postID
	return "Simulated Zipf distribution with enhanced posting and re-posting."
}


// Function to Write JSON Response
func writeResponse(w http.ResponseWriter, response string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": response})
}


// Function to Start the Server
func main() {
	engine := NewEngine()

	http.HandleFunc("/register", enableCORS(engine.RegisterUserHandler)) //POST http://localhost:8080/register?username=user1&password=pass123
	http.HandleFunc("/create_subreddit", enableCORS(engine.CreateSubredditHandler)) //POST http://localhost:8080/create_subreddit?name=golang&creator=user1
	http.HandleFunc("/create_post", enableCORS(engine.CreatePostHandler)) //POST http://localhost:8080/create_post?subreddit=golang&author=user1&content=HelloWorld
	http.HandleFunc("/add_comment", enableCORS(engine.AddCommentHandler)) //POST http://localhost:8080/add_comment?postID=1&author=user1&content=NicePost
	http.HandleFunc("/reply_to_comment", enableCORS(engine.ReplyToCommentHandler)) //POST http://localhost:8080/reply_to_comment?commentID=1&author=user1&content=heey
	http.HandleFunc("/get_feed", enableCORS(engine.GetFeedHandler)) //GET http://localhost:8080/get_feed?subreddit=golang
	http.HandleFunc("/vote_post", enableCORS(engine.VotePostHandler)) //POST http://localhost:8080/vote_post?username=user1&postID=1&vote=1
	http.HandleFunc("/send_message", enableCORS(engine.SendMessageHandler)) //POST http://localhost:8080/send_message?sender=user1&recipient=user1&content=Hey there!
	http.HandleFunc("/list_messages", enableCORS(engine.ListMessagesHandler)) //GET http://localhost:8080/list_messages?username=user1
	http.HandleFunc("/simulate_connection", enableCORS(engine.SimulateConnectionHandler)) //POST http://localhost:8080/simulate_connection?username=user1&connected=true
	http.HandleFunc("/simulate_zipf_distribution", enableCORS(engine.SimulateZipfDistributionHandler)) //POST http://localhost:8080/simulate_zipf_distribution

	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
