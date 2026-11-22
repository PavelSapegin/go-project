package repository
import (
	"context"
	domain "task3-http-grpc/internal/domain"
	redis "github.com/redis/go-redis/v9"
	"strconv"
	"encoding/json"
	"time"
	"errors"
)

type RedisMovieRepository struct {
	client *redis.Client
}

const indexPrefix = "idx:word:"
const cachePrefix = "cache:search:"

func NewRedisMovieRepository(client *redis.Client) *RedisMovieRepository {
	return &RedisMovieRepository{client: client}
}

func (rdb *RedisMovieRepository) IndexMovies(ctx context.Context, movies []*domain.Movie) error {
	if len(movies) == 0 {
		return nil
	}
	
	pipe := rdb.client.Pipeline()

	for _, movie := range movies {
		word1 := domain.Tokenize(movie.Title)
		word2 := domain.Tokenize(movie.Director)
		words := append(word1, word2...)

		for _, word := range words {
			key := indexPrefix + word
			pipe.SAdd(ctx, key, movie.MovieId)
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (rdb *RedisMovieRepository) SearchIDsByWords(ctx context.Context, keywords []string) ([]int, error) {
	if len(keywords) == 0 {
		return []int{}, nil
	}

	keys := make([]string, len(keywords))
	for i, keyword := range keywords {
		keys[i] = indexPrefix + keyword
	}
	
	result, err := rdb.client.SInter(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(result))
	for i, idStr := range result {
		ids[i], err = strconv.Atoi(idStr)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func (rdb *RedisMovieRepository) GetSearchResult(ctx context.Context, key string) ([]*domain.Movie, error) {
	result, err := rdb.client.Get(ctx, cachePrefix+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, domain.ErrCacheMiss
	} else if err != nil {
		return nil, err
	}

	movies := make([]*domain.Movie, 0)
	if err := json.Unmarshal(result, &movies); err != nil {
		return nil, err
	}

	return movies, nil
	
}

func (rdb *RedisMovieRepository) SaveSearchResult(ctx context.Context, key string, movies []*domain.Movie, ttl time.Duration) error {
	if movies == nil {
		movies = []*domain.Movie{}
	}
	
	data, err := json.Marshal(movies)
	if err != nil {
		return err
	}

	err = rdb.client.Set(ctx, cachePrefix+key, data, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rdb *RedisMovieRepository) CleanCache(ctx context.Context) error {
	iter := rdb.client.Scan(ctx, 0, cachePrefix+"*", 0).Iterator()
	keysToDelete := make([]string, 0)
	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
	}


	if err := iter.Err(); err != nil {
		return err
	}

	if len(keysToDelete) > 0 {
		if err := rdb.client.Del(ctx, keysToDelete...).Err(); err != nil {
			return err
		}
	}

	return nil
}
