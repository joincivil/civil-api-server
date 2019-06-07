package posts

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joincivil/civil-api-server/pkg/payments"
	paginator "github.com/pilagod/gorm-cursor-paginator"
	uuid "github.com/satori/go.uuid"
)

// DBPostPersister implements PostPersister interface using Gorm for database persistence
type DBPostPersister struct {
	db *gorm.DB
}

// NewDBPostPersister builds a new DbPostPersister
func NewDBPostPersister(db *gorm.DB) *DBPostPersister {
	return &DBPostPersister{
		db,
	}
}

// CreatePost creates a new Post and saves it to the database
func (p *DBPostPersister) CreatePost(post Post) (Post, error) {
	base, err := PostInterfaceToBase(post)
	if err != nil {
		return nil, err
	}
	if err = p.db.Create(base).Error; err != nil {
		fmt.Printf("An error occured: %v\n", err)
		return nil, err
	}
	return BaseToPostInterface(base)
}

// GetPost retrieves a Post by the id
func (p *DBPostPersister) GetPost(id string) (Post, error) {
	postModel := &PostModel{ID: id}
	p.db.First(postModel)

	return BaseToPostInterface(postModel)
}

func (p *DBPostPersister) GetPayments(post Post) ([]payments.Payment, error) {
	var pays []payments.PaymentModel
	postModel := &PostModel{ID: post.GetPostModel().ID}

	if err := p.db.First(postModel).Related(&pays, "PostPayments").Error; err != nil {
		fmt.Printf("An error occured: %v\n", err)
		return nil, err
	}

	var paymentsSlice []payments.Payment
	for _, result := range pays {
		payment, err := payments.ModelToInterface(&result)
		if err != nil {
			fmt.Printf("An error occured: %v\n", err)
			return nil, err
		}
		paymentsSlice = append(paymentsSlice, payment)
	}

	return paymentsSlice, nil
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

	// stmt.Order("created_at desc").Find(&dbResults)

	results := pager.Paginate(stmt, &dbResults)
	if results.Error != nil {
		fmt.Printf("An error occured: %v\n", results.Error)
		return nil, results.Error
	}

	var posts []Post
	for _, result := range dbResults {
		post, err := BaseToPostInterface(&result)
		if err != nil {
			fmt.Printf("An error occured: %v\n", err)
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
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	base := post.GetPostModel()
	base.ID = id.String()
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
	// TODO(dankins): this should probably use reflection?
	switch base.PostType {
	case "boost":
		post = &Boost{
			PostModel: *base,
		}
	case "externallink":
		post = &ExternalLink{
			PostModel: *base,
		}
	case "comment":
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
