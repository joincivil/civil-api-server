package posts

import (
	"encoding/json"
	"errors"
	"time"

	log "github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	paginator "github.com/pilagod/gorm-cursor-paginator"
	uuid "github.com/satori/go.uuid"
)

var (
	// ErrorNotFound is thrown when trying to retrieve a Post that does not exist
	ErrorNotFound = errors.New("could not find post")
	// ErrorNotAuthorized is thrown when trying to edit a post that you do not have access to
	ErrorNotAuthorized = errors.New("not authorized to perform this action")
	// ErrorNotImplemented is thrown when something isn't implemented
	ErrorNotImplemented = errors.New("not implemented")
	// ErrorBadURLSubmitted is thrown when bad URL is submitted (such as external link that doesn't match newsroom URL)
	ErrorBadURLSubmitted = errors.New("bad URL Submitted")
	// ErrorBadBoostEndDate is thrown when bad URL is submitted (such as external link that doesn't match newsroom URL)
	ErrorBadBoostEndDate = errors.New("bad End Date submitting for Boost")
)

// DBPostPersister implements PostPersister interface using Gorm for database persistence
type DBPostPersister struct {
	db *gorm.DB
}

// NewDBPostPersister builds a new DbPostPersister
func NewDBPostPersister(db *gorm.DB) PostPersister {
	return &DBPostPersister{
		db,
	}
}

// CreatePost creates a new Post and saves it to the database
func (p *DBPostPersister) CreatePost(authorID string, post Post) (Post, error) {
	base, err := PostInterfaceToBase(post)
	if err != nil {
		return nil, err
	}
	if post.GetType() == "boost" {
		boost := post.(*Boost)
		if boost.DateEnd.Before(time.Now()) {
			return nil, ErrorBadBoostEndDate
		}
	}
	base.AuthorID = authorID
	if err = p.db.Create(base).Error; err != nil {
		log.Errorf("An error occurred: %v\n", err)
		return nil, err
	}
	return BaseToPostInterface(base)
}

// GetPost retrieves a Post by the id
func (p *DBPostPersister) GetPost(id string) (Post, error) {
	if id == "" {
		return nil, ErrorNotFound
	}
	postModel := &PostModel{ID: id}
	p.db.First(postModel)

	if (postModel.CreatedAt == time.Time{}) {
		return nil, ErrorNotFound
	}

	return BaseToPostInterface(postModel)
}

// EditPost removes a post from the database (soft_deletes by setting deleted_at flag)
func (p *DBPostPersister) EditPost(requestorUserID string, postID string, patch Post) (Post, error) {

	// get the Post from the database
	dbPost, err := p.GetPost(postID)
	if err != nil {
		return nil, ErrorNotFound
	}

	// check that the requestor has permissions to edit this post
	// TODO(dankins): make this smarter once channel permissions are implemented
	if dbPost.GetPostModel().AuthorID != requestorUserID {
		return nil, ErrorNotAuthorized
	}

	// turn the patch into a JSON object
	jsonPatch, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	// apply the JSON object over the existing post
	// this is a shallow copy
	err = json.Unmarshal(jsonPatch, dbPost)
	if err != nil {
		log.Errorf("error applying patch: %v", err)
		return nil, err
	}

	// Get the fields back into JSON
	jsonPost, err := json.Marshal(dbPost)
	if err != nil {
		return nil, err
	}
	jsonData := json.RawMessage(jsonPost)
	patched := postgres.Jsonb{RawMessage: jsonData}

	// update the Post with the new patched JSON object
	if err = p.db.Model(&PostModel{ID: postID}).Update("data", patched).Error; err != nil {
		log.Errorf("error updating post: %v", err)
		return nil, err
	}

	return dbPost, nil
}

// DeletePost removes a post from the database (soft_deletes by setting deleted_at flag)
func (p *DBPostPersister) DeletePost(requestorUserID string, id string) error {

	if err := p.db.Delete(&PostModel{ID: id}).Error; err != nil {
		log.Errorf("error updating post: %v", err)
		return err
	}

	return nil
}

// SearchPosts retrieves posts making the search criteria
func (p *DBPostPersister) SearchPosts(search *SearchInput) (*PostSearchResult, error) {
	var dbResults []PostModel
	pager := initModelPaginatorFrom(search.Paging)
	stmt := p.db

	if search.PostType != "" {
		stmt = stmt.Where("post_type = ?", search.PostType)
	}
	if search.ChannelID != "" {
		stmt = stmt.Where("channel_id = ?", search.ChannelID)
	}

	results := pager.Paginate(stmt, &dbResults)
	if results.Error != nil {
		log.Errorf("An error occurred: %v\n", results.Error)
		return nil, results.Error
	}

	var posts []Post
	for _, result := range dbResults {
		post, err := BaseToPostInterface(&result)
		if err != nil {
			log.Errorf("An error occurred: %v\n", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	cursors := pager.GetNextCursors()

	response := &PostSearchResult{Posts: posts, Pagination: Pagination{AfterCursor: cursors.AfterCursor, BeforeCursor: cursors.BeforeCursor}}

	return response, nil

}

// initModelPaginatorFrom builds a new paginator
func initModelPaginatorFrom(page Paging) paginator.Paginator {
	p := paginator.New()

	p.SetKeys("CreatedAt", "ID") // [default: "ID"] (order of keys matters)

	if page.AfterCursor != nil {
		p.SetAfterCursor(*page.AfterCursor) // [default: ""]
	}

	if page.BeforeCursor != nil {
		p.SetBeforeCursor(*page.BeforeCursor) // [default: ""]
	}

	if page.Limit != nil {
		p.SetLimit(*page.Limit) // [default: 10]
	}

	if page.Order != nil {
		if *page.Order == "ascending" {
			p.SetOrder(paginator.ASC) // [default: paginator.DESC]
		}
	}
	return p
}

// PostInterfaceToBase takes a post and turns it into a Base ready to go in the database
func PostInterfaceToBase(post Post) (*PostModel, error) {

	base := post.GetPostModel()

	if base.ID == "" {
		id, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		base.ID = id.String()
	}

	base.PostType = post.GetType()

	// turn the post into JSON so we can populate the "data" jsonb column
	// the post struct needs to have Base `json:"-"` otherwise this will pull in all of those fields
	jsonPost, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}
	jsonData := json.RawMessage(jsonPost)
	base.Data = postgres.Jsonb{RawMessage: jsonData}

	return base, nil
}

// BaseToPostInterface accepts a database Base and returns a Post object
func BaseToPostInterface(base *PostModel) (Post, error) {
	var post Post
	switch base.PostType {
	case boost:
		post = &Boost{
			PostModel: *base,
		}
	case externallink:
		post = &ExternalLink{
			PostModel: *base,
		}
	case comment:
		post = &Comment{
			PostModel: *base,
		}
	}
	err := json.Unmarshal(base.Data.RawMessage, post)
	if err != nil {
		return nil, err
	}
	return post.(Post), nil
}
