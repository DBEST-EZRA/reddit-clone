# Reddit Clone Engine and Simulator

A simplified Reddit clone engine implemented in Go, featuring user registration, subreddit management, hierarchical commenting, direct messaging, voting with karma, and simulation of subreddit activity using a Zipf distribution.

## Features

- **User Management**:
  - Register users with unique usernames.
  - Track user karma based on votes on their posts and comments.
  
- **Subreddit Management**:
  - Create and join subreddits.
  - Leave subreddits and view subreddit members.

- **Posts and Comments**:
  - Create posts within subreddits.
  - Add hierarchical comments (nested replies) to posts.
  - Support reposting content across subreddits with reduced initial votes.

- **Voting System**:
  - Upvote or downvote posts.
  - Adjust user karma dynamically based on votes received.

- **Direct Messaging**:
  - Send direct messages between users.
  - View and reply to messages.

- **Simulation Features**:
  - Simulate user connection and disconnection states.
  - Generate subreddit memberships and activity using a Zipf distribution.
  - Create more posts in popular subreddits and introduce reposting.

## Installation and Usage

### Prerequisites

1. Install the Go programming language. Download it from [https://golang.org/dl/](https://golang.org/dl/).

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/DBEST-EZRA/reddit-clone.git
   cd reddit-clone
