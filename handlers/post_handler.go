package handlers

import (
	"blogklert/db"
	"blogklert/middleware"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AppError struct {
	Message string `json:"message"`
	Code    int    `json:"-"`
}

type Post struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Excerpt   string     `json:"excerpt"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

var redisClient *redis.Client

func init() {
	// Initialize Redis client
	redisClient = db.GetRedisClient()
	if redisClient == nil {
		log.Fatal("Failed to initialize Redis client")
	}
}

func SetupPostRoutes(r *mux.Router) {
	r.HandleFunc("/posts", GetPosts).Methods("GET")
	r.HandleFunc("/posts", GetPost).Methods("GET").Queries("id", "{id}")
	r.HandleFunc("/posts", CreatePost).Methods("POST")
	r.HandleFunc("/posts", UpdatePost).Methods("PUT").Queries("id", "{id}")
	r.HandleFunc("/posts", DeletePost).Methods("DELETE").Queries("id", "{id}")
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	// Get the value of the "id" query parameter
	id := r.URL.Query().Get("id")
	if id != "" {
		// If "id" is provided, call GetPost handler with the provided id
		GetPost(w, r)
		return
	}

	// If "id" is not provided, continue with fetching all posts
	ctx := r.Context()

	// Fetch data from cache or database
	posts, err := fetchPosts(ctx)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Marshal the posts data with indentation
	jsonData, err := json.MarshalIndent(posts, "", "    ")
	if err != nil {
		log.Printf("Error marshalling posts data: %v", err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Set Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	if _, err := w.Write(jsonData); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
}

func fetchPosts(ctx context.Context) ([]Post, error) {
	// Check if the data is cached in Redis
	cachedData, err := redisClient.Get(ctx, "posts").Result()
	if err == nil {
		var posts []Post
		if err := json.Unmarshal([]byte(cachedData), &posts); err != nil {
			return nil, fmt.Errorf("error unmarshalling cached posts data: %w", err)
		}
		return posts, nil
	} else if !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("error fetching posts from Redis cache: %w", err)
	}

	// Data not found in cache, fetch from the database
	rows, err := db.DB.QueryContext(ctx, "SELECT * FROM posts")
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Error closing rows: %v", closeErr)
		}
	}()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Excerpt, &post.Body, &post.CreatedAt, &post.UpdatedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	// Cache the fetched data in Redis
	jsonData, err := json.Marshal(posts)
	if err != nil {
		log.Printf("Error marshalling posts data: %v", err)
	} else {
		// CacheTime set to 7 days (604800000 Millisecond)
		const CacheTime = 7 * 24 * time.Hour / time.Millisecond
		if err := redisClient.Set(ctx, "posts", jsonData, CacheTime).Err(); err != nil {
			log.Printf("Error caching posts data: %v", err)
		}
	}

	return posts, nil
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	// Get the value of the "id" query parameter
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		// If "id" is not provided, return a bad request response
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Fetch data from cache or database
	post, err := fetchPost(ctx, idStr)
	if err != nil {
		log.Printf("Error fetching post: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Marshal the post data with indentation
	jsonData, err := json.MarshalIndent(post, "", "    ")
	if err != nil {
		log.Printf("Error marshalling post data: %v", err)
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	// Set Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	if _, err := w.Write(jsonData); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}
}

func fetchPost(ctx context.Context, postID string) (Post, error) {
	// Check if the data is cached in Redis
	cachedData, err := redisClient.Get(ctx, "post:"+postID).Result()
	if err == nil {
		var post Post
		if err := json.Unmarshal([]byte(cachedData), &post); err != nil {
			return Post{}, fmt.Errorf("error unmarshalling cached post data: %w", err)
		}
		return post, nil
	}
	if !errors.Is(err, redis.Nil) {
		return Post{}, fmt.Errorf("error fetching post %s from Redis cache: %w", postID, err)
	}

	// Data not found in cache, fetch from the database
	var post Post
	err = db.DB.QueryRowContext(ctx, "SELECT * FROM posts WHERE id = $1", postID).
		Scan(&post.ID, &post.Title, &post.Excerpt, &post.Body, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Post{}, fmt.Errorf("post %s not found: %w", postID, sql.ErrNoRows)
		}
		return Post{}, fmt.Errorf("error querying database: %w", err)
	}

	// Cache the fetched data in Redis
	jsonData, err := json.Marshal(post)
	if err != nil {
		log.Printf("Error marshalling post data: %v", err)
	} else {
		// CacheTime set to 7 days (604800000 Millisecond)
		const CacheTime = 7 * 24 * time.Hour / time.Millisecond
		if err := redisClient.Set(ctx, "post:"+postID, jsonData, CacheTime).Err(); err != nil {
			log.Printf("Error caching post %s data: %v", postID, err)
		}
	}

	return post, nil
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode JSON payload directly into a Post struct
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Sanitize post fields
	post.Title = middleware.SanitizeInput(post.Title, 15)
	post.Excerpt = middleware.SanitizeInput(post.Excerpt, 150)
	post.Body = middleware.SanitizeInput(post.Body, 1500)

	// Validate post fields
	if err := post.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert the post into the database
	if err := insertPost(post); err != nil {
		log.Printf("Error inserting post: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Invalidate the cache for the list of all posts
	if err := redisClient.Del(ctx, "posts").Err(); err != nil {
		log.Printf("Error invalidating cache for posts: %v", err)
	}

	respondJSON(w, nil, http.StatusCreated)
}

func insertPost(post Post) error {
	_, err := db.DB.Exec("INSERT INTO posts (title, excerpt, body) VALUES ($1, $2, $3)",
		post.Title, post.Excerpt, post.Body)
	if err != nil {
		return fmt.Errorf("failed to insert post: %w", err)
	}
	return nil
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract the post ID from query parameters
	postID := r.URL.Query().Get("id")
	if postID == "" {
		log.Println("Post ID is required")
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Parse post ID to UUID
	id, err := uuid.Parse(postID)
	if err != nil {
		log.Printf("Error parsing ID to UUID: %v", err)
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	// Validate content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	// Decode JSON payload directly into a Post struct
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Sanitize post fields
	post.Title = middleware.SanitizeInput(post.Title, 15)
	post.Excerpt = middleware.SanitizeInput(post.Excerpt, 150)
	post.Body = middleware.SanitizeInput(post.Body, 1500)

	// Validate post fields
	if err := post.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the ID of the post
	post.ID = id

	// Update the post
	if err := updatePost(post); err != nil {
		log.Printf("Error updating post: %v", err)
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}

	// Invalidate the cache for both the updated post and the list of all posts
	postCacheKey := "post:" + postID
	if err := redisClient.Del(ctx, postCacheKey).Err(); err != nil {
		log.Printf("Error invalidating cache for post %s: %v", postID, err)
	}
	if err := redisClient.Del(ctx, "posts").Err(); err != nil {
		log.Printf("Error invalidating cache for posts: %v", err)
	}

	respondJSON(w, nil, http.StatusNoContent)
}

func updatePost(post Post) error {
	_, err := db.DB.Exec("UPDATE posts SET title = $1, excerpt = $2, body = $3 WHERE id = $4",
		post.Title, post.Excerpt, post.Body, post.ID)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queryParams := r.URL.Query()
	postID := queryParams.Get("id")

	// Parse postID to UUID
	id, err := uuid.Parse(postID)
	if err != nil {
		log.Printf("Error parsing ID to UUID: %v", err)
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := deletePost(id); err != nil {
		log.Printf("Error deleting post: %v", err)
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	// Invalidate the cache for both the deleted post and the list of all posts
	postCacheKey := "post:" + postID
	if err := redisClient.Del(ctx, postCacheKey).Err(); err != nil {
		log.Printf("Error invalidating cache for post %s: %v", postID, err)
	}
	if err := redisClient.Del(ctx, "posts").Err(); err != nil {
		log.Printf("Error invalidating cache for posts: %v", err)
	}

	respondJSON(w, nil, http.StatusNoContent)
}

func deletePost(id uuid.UUID) error {
	_, err := db.DB.Exec("DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

func (p *Post) Validate() error {
	if p.Title == "" {
		return errors.New("title cannot be empty")
	}
	if wordCount(p.Title) > 15 {
		return errors.New("title exceeds maximum word count of 15")
	}
	if p.Excerpt == "" {
		return errors.New("excerpt cannot be empty")
	}
	if wordCount(p.Excerpt) > 150 {
		return errors.New("excerpt exceeds maximum word count of 150")
	}
	if p.Body == "" {
		return errors.New("body cannot be empty")
	}
	if wordCount(p.Body) > 1500 {
		return errors.New("body exceeds maximum word count of 1500")
	}
	return nil
}

func wordCount(s string) int {
	words := strings.Fields(s)
	return len(words)
}

func (e *AppError) Error() string {
	return e.Message
}

func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON: %v", err)
		}
	}
}
