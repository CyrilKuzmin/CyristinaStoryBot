package story

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dirName = "./stories"

type Story struct {
	ID      int64
	Title   string
	Content []ContentPart
	Parts   int
}

type ContentPart struct {
	Image   string
	Caption string
}

type StoryMongoClient struct {
	Client     *mongo.Client
	DB         string
	Collection string
}

func NewClient(uri, db, collection string) (StoryMongoClient, error) {
	var client StoryMongoClient
	var err error
	clientOptions := options.Client().ApplyURI(uri)
	client.Client, err = mongo.Connect(context.TODO(), clientOptions)
	client.DB = db
	client.Collection = collection
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

func (cliient StoryMongoClient) GetStoryPart(title string, part int) (ContentPart, error) {
	var cp ContentPart
	st, err := cliient.getStoryByTitle(title)
	if err != nil {
		return cp, err
	}
	cp = st.Content[part]
	return cp, nil
}

func (client StoryMongoClient) GetAllTitles() ([]string, error) {
	type titObj struct {
		Title string
	}

	var titles []titObj
	var result []string

	filter := bson.D{{}}                 // Все документы
	selectFilter := bson.D{{"Title", 1}} // Только Title

	col := client.Client.Database(client.DB).Collection(client.Collection)
	cur, err := col.Find(context.TODO(), filter, options.Find().SetProjection(selectFilter))
	if err != nil {
		return result, err
	}
	defer cur.Close(context.TODO())

	cur.All(context.TODO(), &titles)
	if err := cur.Err(); err != nil {
		return result, err
	}

	for _, t := range titles {
		result = append(result, t.Title)
	}
	return result, nil
}
