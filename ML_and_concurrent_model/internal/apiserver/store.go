package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ---- MongoDB document types ----

// JobDoc tracks the lifecycle of a distributed training job.
type JobDoc struct {
	ID          bson.ObjectID  `bson:"_id,omitempty"    json:"id"`
	JobID       string         `bson:"job_id"           json:"job_id"`
	Status      string         `bson:"status"           json:"status"` // pending|running|done|failed
	Progress    int            `bson:"progress"         json:"progress"`
	StartedAt   time.Time      `bson:"started_at"       json:"started_at"`
	FinishedAt  *time.Time     `bson:"finished_at"      json:"finished_at,omitempty"`
	ErrorMsg    string         `bson:"error,omitempty"  json:"error,omitempty"`
	ModelID     string         `bson:"model_id"         json:"model_id,omitempty"`
}

// ModelDoc stores a trained model: coefficients, metrics, and all the metadata
// needed to reproduce predictions without reloading the dataset.
type ModelDoc struct {
	ID              bson.ObjectID `bson:"_id,omitempty"      json:"id"`
	JobID           string        `bson:"job_id"             json:"job_id"`
	Beta            []float64     `bson:"beta"               json:"beta"`
	MAE             float64       `bson:"mae"                json:"mae"`
	RMSE            float64       `bson:"rmse"               json:"rmse"`
	R2              float64       `bson:"r2"                 json:"r2"`
	TrainedAt       time.Time     `bson:"trained_at"         json:"trained_at"`
	Solver          string        `bson:"solver"             json:"solver"`
	Lambda          float64       `bson:"lambda"             json:"lambda"`
	NodesUsed       []string      `bson:"nodes_used"         json:"nodes_used"`
	FeatureNames    []string      `bson:"feature_names"      json:"feature_names"`
	NumFeatureNames []string      `bson:"num_feature_names"  json:"num_feature_names"`
	CatFeatureNames []string      `bson:"cat_feature_names"  json:"cat_feature_names"`
	Means           []float64     `bson:"means"              json:"means"`
	Stds            []float64     `bson:"stds"               json:"stds"`
	PlanJSON        []byte        `bson:"plan_json"          json:"-"`
	Cols            int           `bson:"cols"               json:"cols"`
	TrainRows       int           `bson:"train_rows"         json:"train_rows"`
	TestRows        int           `bson:"test_rows"          json:"test_rows"`
}

// ---- Store ----

// Store holds both storage backends.  Redis is optional; if nil, caching is skipped.
type Store struct {
	mongo *mongo.Client
	rdb   *redis.Client
	db    string
}

// NewStore connects to MongoDB and (optionally) Redis.
// redisAddr may be empty to disable Redis.
func NewStore(mongoURL, dbName, redisAddr string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mc, err := mongo.Connect(options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := mc.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	s := &Store{mongo: mc, db: dbName}

	if redisAddr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
		if err := rdb.Ping(ctx).Err(); err != nil {
			// Redis is non-critical — log and continue without it.
			fmt.Printf("[store] Redis no disponible (%v); cache deshabilitado\n", err)
		} else {
			s.rdb = rdb
		}
	}

	return s, nil
}

func (s *Store) col(name string) *mongo.Collection {
	return s.mongo.Database(s.db).Collection(name)
}

// ---- Job operations ----

func (s *Store) CreateJob(ctx context.Context, doc *JobDoc) error {
	_, err := s.col("jobs").InsertOne(ctx, doc)
	return err
}

func (s *Store) UpdateJob(ctx context.Context, jobID string, update bson.M) error {
	_, err := s.col("jobs").UpdateOne(
		ctx,
		bson.M{"job_id": jobID},
		bson.M{"$set": update},
	)
	return err
}

func (s *Store) GetJob(ctx context.Context, jobID string) (*JobDoc, error) {
	var doc JobDoc
	err := s.col("jobs").FindOne(ctx, bson.M{"job_id": jobID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// ---- Model operations ----

func (s *Store) SaveModel(ctx context.Context, doc *ModelDoc) (string, error) {
	res, err := s.col("models").InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		return oid.Hex(), nil
	}
	return fmt.Sprintf("%v", res.InsertedID), nil
}

func (s *Store) GetModel(ctx context.Context, id string) (*ModelDoc, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid model id: %w", err)
	}
	var doc ModelDoc
	if err := s.col("models").FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *Store) LatestModel(ctx context.Context) (*ModelDoc, error) {
	opts := options.FindOne().SetSort(bson.M{"trained_at": -1})
	var doc ModelDoc
	if err := s.col("models").FindOne(ctx, bson.M{}, opts).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *Store) ListModels(ctx context.Context, limit int) ([]ModelDoc, error) {
	opts := options.Find().SetSort(bson.M{"trained_at": -1}).SetLimit(int64(limit))
	cur, err := s.col("models").Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var docs []ModelDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// ---- Redis cache ----

const predCacheTTL = 24 * time.Hour

// CachePrediction stores a prediction value in Redis keyed by a hash string.
func (s *Store) CachePrediction(ctx context.Context, key string, value float64) {
	if s.rdb == nil {
		return
	}
	data, _ := json.Marshal(value)
	s.rdb.Set(ctx, "pred:"+key, data, predCacheTTL) //nolint:errcheck
}

// GetCachedPrediction retrieves a cached prediction. ok=false if not found.
func (s *Store) GetCachedPrediction(ctx context.Context, key string) (float64, bool) {
	if s.rdb == nil {
		return 0, false
	}
	data, err := s.rdb.Get(ctx, "pred:"+key).Bytes()
	if err != nil {
		return 0, false
	}
	var v float64
	if json.Unmarshal(data, &v) != nil {
		return 0, false
	}
	return v, true
}
