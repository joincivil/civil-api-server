package posts

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	paginator "github.com/pilagod/gorm-cursor-paginator"
	uuid "github.com/satori/go.uuid"
	"time"
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
	// ErrorNoReferenceURLFound is thrown when no canonical URL or OG URL is found for a submitted external link
	ErrorNoReferenceURLFound = errors.New("no canonical URL or Open Graph URL found on submitted page")
	// ErrorBadFilterProvided is thrown when storyfeed cannot be queried due to bad filter input
	ErrorBadFilterProvided = errors.New("bad storyfeed filter provided")
)

const (
	chronologicalExternallinkViewName = "vw_post_externallink_chronological"
	chronologicalBoostViewName        = "vw_post_boost_chronological"
	fairThenChronologicalViewName     = "vw_post_fair_then_chronological_2"
	fairWithInterleavedBoostsViewName = "vw_post_fair_with_interleaved_boosts_2"
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

// CreateViews creates the views if they don't exist
func (p *DBPostPersister) CreateViews() error {
	createStoryfeedViewQuery := CreateChronologicalStoryfeedViewQuery(chronologicalExternallinkViewName)
	db := p.db.Exec(createStoryfeedViewQuery)
	if db.Error != nil {
		return fmt.Errorf("Error creating chronological storyfeed view in postgres: %v", db.Error)
	}
	createBoostfeedViewQuery := CreateChronologicalBoostfeedViewQuery(chronologicalBoostViewName)
	db = p.db.Exec(createBoostfeedViewQuery)
	if db.Error != nil {
		return fmt.Errorf("Error creating chronological boostfeed view in postgres: %v", db.Error)
	}
	createStoryfeedViewQuery = CreateFairThenChronologicalStoryfeedViewQuery(fairThenChronologicalViewName)
	db = p.db.Exec(createStoryfeedViewQuery)
	if db.Error != nil {
		return fmt.Errorf("Error creating fair then chronological storyfeed view in postgres: %v", db.Error)
	}
	createStoryfeedViewQuery = CreateFairWithInterleavedBoostsStoryfeedViewQuery(fairWithInterleavedBoostsViewName)
	db = p.db.Exec(createStoryfeedViewQuery)
	if db.Error != nil {
		return fmt.Errorf("Error creating fair with interleaved boosts storyfeed view in postgres: %v", db.Error)
	}
	return nil
}

// CreateChronologicalStoryfeedViewQuery returns the query to create the storyfeed view
func CreateChronologicalStoryfeedViewQuery(viewName string) string {
	queryString := fmt.Sprintf(`
	CREATE OR REPLACE VIEW %s as (
		select *, coalesce((data ->> 'published_time')::timestamptz, created_at)::timestamptz as sort_date 
		from posts 
		where post_type = 'externallink'
		order by sort_date desc
	)
    `, viewName)
	return queryString
}

func (p *DBPostPersister) getRawChannelChronologicalStoryfeedQuery(channelID string) *gorm.DB {
	return p.db.Raw("select *, coalesce((data ->> 'published_time')::timestamptz, created_at)::timestamptz as sort_date from posts where post_type = 'externallink' and channel_id = ? order by sort_date desc", channelID)
}

// CreateChronologicalBoostfeedViewQuery returns the query to create the boostfeed view
func CreateChronologicalBoostfeedViewQuery(viewName string) string {
	queryString := fmt.Sprintf(`
	CREATE OR REPLACE VIEW %s as (
		select * 
		from posts 
		where post_type = 'boost'
		order by created_at desc
	)
    `, viewName)
	return queryString
}

func (p *DBPostPersister) getRawChannelChronologicalBoostfeedQuery(channelID string) *gorm.DB {
	return p.db.Raw("select *, coalesce((data ->> 'published_time')::timestamptz, created_at)::timestamptz as sort_date from posts where post_type = 'boost' and channel_id = ? order by sort_date desc", channelID)
}

// CreateFairThenChronologicalStoryfeedViewQuery returns the query to create the storyfeed view
func CreateFairThenChronologicalStoryfeedViewQuery(viewName string) string {
	queryString := fmt.Sprintf(`
	CREATE OR REPLACE VIEW %s as (
		select *, (case when post_num = 1 then 1 ELSE null end) as rank  FROM
		(
			select
				*,
				count(1) over (partition by channel_id order by (sort_date) desc) as post_num
				from
				(
				select 
					*, 
					coalesce((data ->> 'published_time')::timestamptz, created_at)::timestamptz as sort_date
					from posts
					where post_type = 'externallink'
				) data2
				
		) data
		order by rank, sort_date desc
	)
    `, viewName)
	return queryString
}

// the "fair the chronological" is identical to "chronological" for a single channel
func (p *DBPostPersister) getRawChannelFairThenChronologicalStoryfeedQuery(channelID string) *gorm.DB {
	return p.db.Raw("select *, coalesce((data ->> 'published_time')::timestamptz, created_at)::timestamptz as sort_date from posts where post_type = 'externallink' and channel_id = ? order by sort_date desc", channelID)
}

// CreateFairWithInterleavedBoostsStoryfeedViewQuery returns the query to create the storyfeed view
func CreateFairWithInterleavedBoostsStoryfeedViewQuery(viewName string) string {
	// nolint: gosec
	queryString := fmt.Sprintf(`
	CREATE OR REPLACE VIEW %s as (
		SELECT * FROM (
			select *, (case when post_num = 1 then 1 ELSE null end) as rank, ROW_NUMBER() OVER (ORDER BY (case when post_num = 1 then 1 ELSE null end), sort_date desc) as row_rank  FROM
			(
				select
					*,
					count(1) over (partition by channel_id order by (sort_date) desc) as post_num
					from
					(
					select 
						*, 
						coalesce((data ->> 'published_time')::timestamptz, created_at)::timestamptz as sort_date
						from posts
						where post_type = 'externallink'
					) data2
					
			) data
			order by row_rank desc
		) data3

		UNION

		SELECT * FROM
		(
			SELECT *, 1 as rank, ROW_NUMBER() OVER (ORDER BY sort_date) * 5 as row_rank FROM
			(
				SELECT *, created_at as sort_date, 1 as post_num FROM posts where post_type = 'boost' and (data ->> 'date_end')::timestamp > now()
				
			) data2
			order by sort_date desc
		) data4
		order by row_rank
	)
    `, viewName)
	return queryString
}

func (p *DBPostPersister) getRawChannelFairWithInterleavedBoostsStoryfeedQuery(channelID string) *gorm.DB {
	return p.db.Raw(`
		SELECT * FROM (
			SELECT *, (CASE WHEN post_num = 1 THEN 1 ELSE NULL END) AS rank, ROW_NUMBER() OVER (ORDER BY (CASE WHEN post_num = 1 THEN 1 ELSE NULL END), sort_date desc) AS row_rank  FROM
			(
				SELECT
					*,
					COUNT(1) OVER (PARTITION BY channel_id order by (sort_date) desc) AS post_num
					FROM
					(
					SELECT 
						*, 
						coalesce((data ->> 'published_time')::timestamptz, created_at)::timestamptz AS sort_date
						FROM posts
						WHERE post_type = 'externallink'
						AND channel_id = ?
					) data2
					
			) data
			ORDER BY row_rank desc
		) data3

		UNION

		SELECT * FROM
		(
			SELECT *, 1 AS rank, ROW_NUMBER() OVER (ORDER BY sort_date) * 3 AS row_rank FROM
			(
				SELECT *, created_at AS sort_date, 1 AS post_num
				FROM posts
				WHERE post_type = 'boost' AND
				(data ->> 'date_end')::timestamp > now() AND
				channel_id = ?
				
			) data2
			ORDER BY sort_date desc
		) data4
		ORDER BY row_rank`, channelID, channelID)
}

// CreatePost creates a new Post and saves it to the database
func (p *DBPostPersister) CreatePost(authorID string, post Post) (Post, error) {
	base, err := PostInterfaceToBase(post)
	if err != nil {
		return nil, err
	}
	if post.GetType() == TypeBoost {
		boost := post.(Boost)
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

// GetPostByReference retrieves a Post by the reference
func (p *DBPostPersister) GetPostByReference(reference string) (Post, error) {
	if reference == "" {
		return nil, ErrorNotFound
	}
	postModel := &PostModel{}
	p.db.Where("reference  = ?", reference).First(postModel)

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

func (p *DBPostPersister) getRawStoryfeedQuery(limit int, offset int, storyfeedViewName string, channelID *string) *gorm.DB {
	if channelID == nil {
		return p.db.Raw(fmt.Sprintf("select * from %s limit %d offset %d", storyfeedViewName, limit, offset))
	}

	switch storyfeedViewName {
	case chronologicalExternallinkViewName:
		return p.getRawChannelChronologicalStoryfeedQuery(*channelID)
	case chronologicalBoostViewName:
		return p.getRawChannelChronologicalBoostfeedQuery(*channelID)
	case fairThenChronologicalViewName:
		return p.getRawChannelFairThenChronologicalStoryfeedQuery(*channelID)
	case fairWithInterleavedBoostsViewName:
		return p.getRawChannelFairWithInterleavedBoostsStoryfeedQuery(*channelID)
	}

	return nil
}

// SearchPostsRanked retrieves most recent externallink post for each channel, followed by all the rest of the posts in reverse chronological order
func (p *DBPostPersister) SearchPostsRanked(limit int, offset int, filter *StoryfeedFilter) (*PostSearchResult, error) {
	var dbResults []PostModel

	var channelID *string
	storyfeedViewName := fairThenChronologicalViewName // backwards compatible for queries that don't include filter
	if filter != nil {
		storyfeedViewName = filter.Alg
		channelID = filter.ChannelID
	}

	stmt := p.getRawStoryfeedQuery(limit, offset, storyfeedViewName, channelID)
	if stmt == nil {
		return nil, ErrorBadFilterProvided
	}

	results := stmt.Scan(&dbResults)
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

	response := &PostSearchResult{Posts: posts}

	return response, nil
}

// SearchPostsMostRecentPerChannel retrieves most recent post for each channel matching the search criteria
func (p *DBPostPersister) SearchPostsMostRecentPerChannel(search *SearchInput) (*PostSearchResult, error) {
	var dbResults []PostModel

	pager := initModelPaginatorFrom(search.Paging)
	stmt := p.db.Where("created_at IN(SELECT MAX(created_at) FROM posts WHERE deleted_at IS NULL GROUP BY channel_id)", search.PostType)
	if search.PostType != "" {
		stmt = p.db.Where("created_at IN(SELECT MAX(created_at) FROM posts WHERE deleted_at IS NULL AND post_type = ? GROUP BY channel_id)", search.PostType)
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

// SearchPosts retrieves posts matching the search criteria
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
		id := uuid.NewV4()
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
	case TypeBoost:
		post = &Boost{
			PostModel: *base,
		}
	case TypeExternalLink:
		post = &ExternalLink{
			PostModel: *base,
		}
	case TypeComment:
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
