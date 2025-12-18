package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:150;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;size:254;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	FirstName string    `gorm:"size:150" json:"first_name"`
	LastName  string    `gorm:"size:150" json:"last_name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Отношения
	Profile    Profile     `gorm:"foreignKey:UserID" json:"profile"`
	Stories    []Story     `gorm:"foreignKey:UserID" json:"stories,omitempty"`
	Comments   []Comment   `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Likes      []Like      `gorm:"foreignKey:UserID" json:"likes,omitempty"`
	Followers  []Subscription `gorm:"foreignKey:FollowingID" json:"followers,omitempty"`
	Following  []Subscription `gorm:"foreignKey:FollowerID" json:"following,omitempty"`
}

type Profile struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Avatar       string    `gorm:"size:500" json:"avatar"`
	Bio          string    `gorm:"type:text" json:"bio"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	IsEarly      bool      `gorm:"default:false" json:"is_early"`
	OtpCode      string    `gorm:"size:6" json:"-"`
	OtpCreatedAt time.Time `json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	StreakCount      int       `gorm:"default:0" json:"streak_count"`
	LastActiveAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"last_active_at"`
    StreakRewarded   bool      `gorm:"default:false" json:"streak_rewarded"`
}

// models/story.go - ДОБАВИТЬ ПОЛЯ:
type Story struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	
	// НОВЫЕ ПОЛЯ ДЛЯ RAVELL
	WordCount  int        `gorm:"default:0" json:"word_count"`      // всегда 100
	ReplyTo    *uint      `gorm:"index" json:"reply_to"`            // null для корня
	ReplyCount int        `gorm:"default:0" json:"reply_count"`     // количество ответов
	LastReplyAt *time.Time `gorm:"index" json:"last_reply_at"`      // время последнего ответа
	
	// Существующие поля
	ImageURL  string    `gorm:"size:500" json:"image_url"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Отношения
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	Comments  []Comment      `gorm:"foreignKey:StoryID" json:"comments,omitempty"`
	Likes     []Like         `gorm:"foreignKey:StoryID" json:"likes,omitempty"`
	Hashtags  []StoryHashtag `gorm:"foreignKey:StoryID" json:"hashtags,omitempty"`
}
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	StoryID   uint      `gorm:"not null;index" json:"story_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	
	// Отношения
	User  User  `gorm:"foreignKey:UserID" json:"user"`
	Story Story `gorm:"foreignKey:StoryID" json:"story"`
}

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_story" json:"user_id"`
	StoryID   uint      `gorm:"not null;uniqueIndex:idx_user_story" json:"story_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Subscription struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FollowerID  uint      `gorm:"not null;uniqueIndex:idx_follower_following" json:"follower_id"`
	FollowingID uint      `gorm:"not null;uniqueIndex:idx_follower_following" json:"following_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Hashtag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	
	// Отношения
	Stories []StoryHashtag `gorm:"foreignKey:HashtagID" json:"stories,omitempty"`
}

type StoryHashtag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StoryID   uint      `gorm:"not null;uniqueIndex:idx_story_hashtag" json:"story_id"`
	HashtagID uint      `gorm:"not null;uniqueIndex:idx_story_hashtag" json:"hashtag_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type NotInterested struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_story" json:"user_id"`
	StoryID   uint      `gorm:"not null;uniqueIndex:idx_user_story" json:"story_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type UserDevice struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    PlayerID  string    `gorm:"not null" json:"player_id"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

    User User `gorm:"foreignKey:UserID" json:"user"`
}

type Feature struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"` // кто предложил
    Title     string    `gorm:"size:255;not null" json:"title"`
    Description string  `gorm:"type:text" json:"description"`
    UsedInRelease bool   `gorm:"default:false" json:"used_in_release"` // фича была использована
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

    User User `gorm:"foreignKey:UserID" json:"user"`
}

type Achievement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"uniqueIndex;size:100;not null" json:"key"` // уникальный ключ, напр. "early_access"
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type UserAchievement struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"not null;index" json:"user_id"`
	AchievementID uint       `gorm:"not null;index" json:"achievement_id"`
	Progress      float64    `gorm:"default:0" json:"progress"` // 0..1
	Unlocked      bool       `gorm:"default:false" json:"unlocked"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	User        User        `gorm:"foreignKey:UserID" json:"user"`
	Achievement Achievement `gorm:"foreignKey:AchievementID" json:"achievement"`
}