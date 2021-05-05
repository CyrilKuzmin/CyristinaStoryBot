package story

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Story struct {
	ID      int64
	Title   string
	Content []ContentPart
}

type ContentPart struct {
	Image   string
	Caption string
}

type StoryMongoClient struct {
	Client      *mongo.Client
	DB          string
	Collection  string
	allTitles   []string
	Titles      map[string]bool
	TitlesCount int
}

func NewClient(uri, db, collection string) (StoryMongoClient, error) {
	var client StoryMongoClient
	var err error
	clientOptions := options.Client().ApplyURI(uri)
	client.Client, err = mongo.Connect(context.TODO(), clientOptions)
	client.DB = db
	client.Collection = collection
	client.Titles = make(map[string]bool)
	if err != nil {
		return client, err
	}
	return client, nil
}

func (client StoryMongoClient) getStoryByTitle(title string) (Story, error) {
	var result Story
	filter := bson.D{{"Title", title}}
	col := client.Client.Database(client.DB).Collection(client.Collection)
	err := col.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (cliient StoryMongoClient) GetStoryPart(
	title string,
	part int) (ContentPart, error) {
	var cp ContentPart
	st, err := cliient.getStoryByTitle(title)
	if err != nil {
		return cp, err
	}
	cp = st.Content[part]
	return cp, nil
}

func (client *StoryMongoClient) GetAllTitles() error {
	client.allTitles = make([]string, 0, 10)

	type titObj struct {
		Title string
	}

	var titles []titObj

	filter := bson.D{{}}                 // Все документы
	selectFilter := bson.D{{"Title", 1}} // Только Title

	col := client.Client.Database(client.DB).Collection(client.Collection)
	cur, err := col.Find(
		context.TODO(),
		filter,
		options.Find().SetProjection(selectFilter))
	if err != nil {
		return err
	}
	defer cur.Close(context.TODO())

	cur.All(context.TODO(), &titles)
	if err := cur.Err(); err != nil {
		return err
	}

	for _, t := range titles {
		client.allTitles = append(client.allTitles, t.Title)
		client.Titles[t.Title] = true
	}

	client.allTitles = append(client.allTitles, "THE_END")
	client.TitlesCount = len(client.allTitles)
	return nil
}

// GetTitlesPart возвращает часть тайтлов (10 штук)
func (client StoryMongoClient) GetTitlesPart(part int) (titles []string) {
	firstTitle := part * 10
	lastTitle := part*10 + 9
	if firstTitle > len(client.allTitles)-1 {
		firstTitle = len(client.allTitles)
	}
	if lastTitle > len(client.allTitles)-1 {
		lastTitle = len(client.allTitles)
	}
	return client.allTitles[firstTitle:lastTitle]
}
