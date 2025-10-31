package storage

import (
    "context"
    "log"
    "os"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
    client     *mongo.Client
    collection *mongo.Collection
}

func NewStorage() *Storage {
    uri := os.Getenv("MONGO_URI")
    if uri == "" {
        uri = "mongodb://localhost:27017"
    }

    dbName := os.Getenv("MONGO_DB")
    if dbName == "" {
        dbName = "fourinrow"
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatalf("❌ MongoDB connection failed: %v", err)
    }

    log.Println("✅ Connected to MongoDB:", uri)

    return &Storage{
        client:     client,
        collection: client.Database(dbName).Collection("leaderboard"),
    }
}

// IncrementWin increases the win count for a player or inserts them if new
func (s *Storage) IncrementWin(username string) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    filter := bson.M{"username": username}
    update := bson.M{"$inc": bson.M{"wins": 1}}
    opts := options.Update().SetUpsert(true)

    _, err := s.collection.UpdateOne(ctx, filter, update, opts)
    if err != nil {
        log.Printf("❌ Failed to update win count for %s: %v", username, err)
    } else {
        log.Printf("🏆 Win recorded for %s", username)
    }
}

// GetLeaderboard retrieves top players
func (s *Storage) GetLeaderboard() []bson.M {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cursor, err := s.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"wins": -1}))
    if err != nil {
        log.Println("❌ Failed to fetch leaderboard:", err)
        return nil
    }
    defer cursor.Close(ctx)

    var results []bson.M
    if err = cursor.All(ctx, &results); err != nil {
        log.Println("❌ Cursor decode error:", err)
    }

    return results
}
