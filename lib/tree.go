package lib

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

// A Node is any node which is part of the post tree.
type Node interface {
	String() string
	IsPost() bool
}

// A Post represents a post in the tree.
type Post struct {
	ID       int
	Content  string
	Score    int
	User     string
	Children []Node
}

func (p *Post) String() string {
	children := "["

	for _, child := range p.Children {
		children += child.String()
	}

	children += "]"

	return fmt.Sprintf(
		"(user:%s  content:'%s'  score:%d  children:%s)",
		p.User,
		p.Content,
		p.Score,
		children,
	)
}

// IsPost returns true if the node is a post. If a node is a post, it's assumed
// that it has an ID, Content, Score, User, and Children.
func (p *Post) IsPost() bool { return true }

// A MaxDepthMarker indicates that the maxDepth was reached when constructing
// the post tree, and contains the id which the node would've been.
type MaxDepthMarker struct {
	ID int
}

func (m *MaxDepthMarker) String() string {
	return fmt.Sprintf("(max depth reached at id:%d)", m.ID)
}

// IsPost returns true if the node is a post. If a node is a post, it's assumed
// that it has an ID, Content, Score, User, and Children.
func (m *MaxDepthMarker) IsPost() bool { return false }

// FetchPostTree creates a tree containing a tree of all the posts under the
// post which has id == head.
func FetchPostTree(head, maxDepth int, rdb *redis.Client) (Node, error) {
	if maxDepth == 0 {
		return &MaxDepthMarker{ID: head}, nil
	}

	var (
		key         = fmt.Sprintf("post:%d", head)
		childrenKey = fmt.Sprintf("%s:children", key)
		p           = new(Post)
	)

	p.ID = head

	content, err := rdb.HGet(key, "content").Result()
	if err != nil {
		return nil, err
	}
	p.Content = content

	score, err := rdb.HGet(key, "score").Result()
	if err != nil {
		return nil, err
	}

	scoreVal, err := strconv.Atoi(score)
	if err != nil {
		return nil, err
	}
	p.Score = scoreVal

	user, err := rdb.HGet(key, "user").Result()
	if err != nil {
		return nil, err
	}
	p.User = user

	children, err := rdb.SMembers(childrenKey).Result()
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		id, err := strconv.Atoi(child)
		if err != nil {
			return nil, err
		}

		subTree, err := FetchPostTree(id, maxDepth-1, rdb)
		if err != nil {
			return nil, err
		}

		if subTree != nil {
			p.Children = append(p.Children, subTree)
		}
	}

	return p, nil
}
