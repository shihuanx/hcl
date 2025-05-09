package user_repository

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"huancuilou/internal/user/user_model"
	"log"
	"strconv"
	"time"
)

type UserCacheRepository struct {
	client *redis.Client
}

// NewUserCacheRepository 初始化缓存层结构体实例
func NewUserCacheRepository(client *redis.Client) *UserCacheRepository {
	return &UserCacheRepository{
		client: client,
	}
}

// UserCachePrefix 定义缓存键的前缀
const UserCachePrefix = "hcl:user"

func (u *UserCacheRepository) AddFollows(follow *user_model.UserFollow) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:follows:%d", UserCachePrefix, follow.UserID)
	err := u.client.SAdd(ctx, key, follow.FollowID).Err()
	if err != nil {
		return err
	}
	log.Printf("%d添加关注%d", follow.UserID, follow.FollowID)
	return nil
}

func (u *UserCacheRepository) AddFans(follow *user_model.UserFollow) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:fans:%d", UserCachePrefix, follow.FollowID)
	err := u.client.SAdd(ctx, key, follow.UserID).Err()
	if err != nil {
		return err
	}
	log.Printf("%d添加粉丝%d", follow.FollowID, follow.UserID)
	return nil
}

func (u *UserCacheRepository) RemoveFollows(userID int, followID int) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:follows:%d", UserCachePrefix, userID)
	err := u.client.SRem(ctx, key, followID).Err()
	if err != nil {
		return err
	}
	return nil
}

func (u *UserCacheRepository) RemoveFans(userID int, followID int) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:fans:%d", UserCachePrefix, followID)
	err := u.client.SRem(ctx, key, userID).Err()
	if err != nil {
		return err
	}
	return nil
}

func (u *UserCacheRepository) GetCommonFollows(userID int, otherUserID int) ([]int, error) {
	ctx := context.Background()
	var commonFollows []int
	followKey := fmt.Sprintf("%s:follows:%d", UserCachePrefix, userID)
	fanKey := fmt.Sprintf("%s:fans:%d", UserCachePrefix, otherUserID)
	commonFollowsStr, err := u.client.SInter(ctx, followKey, fanKey).Result()
	if err != nil {
		return nil, err
	}
	for _, v := range commonFollowsStr {
		commonFollow, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		commonFollows = append(commonFollows, commonFollow)
	}
	return commonFollows, nil
}

func (u *UserCacheRepository) AddLikes(id int, score int) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:likes", UserCachePrefix)
	result := u.client.ZIncrBy(ctx, key, float64(score), strconv.Itoa(id))
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (u *UserCacheRepository) GetLikesRank(userID int) ([]*user_model.UserLikeRank, error) {
	ctx := context.Background()
	// 计算交集并将结果存储到新的有序集合中
	likeRankKey := fmt.Sprintf("%s:likesRank:%d", UserCachePrefix, userID)
	followsKey := fmt.Sprintf("%s:follows:%d", UserCachePrefix, userID)
	allLikesKey := fmt.Sprintf("%s:likes", UserCachePrefix)

	//从点赞有序集合中获得本人的点赞量
	userLikeCount := u.client.ZScore(ctx, allLikesKey, strconv.Itoa(userID))

	//将交集结果存储到新的有序集合中并保留权值
	_, err := u.client.ZInterStore(ctx, likeRankKey, &redis.ZStore{
		Keys:      []string{allLikesKey, followsKey},
		Weights:   []float64{1, 0},
		Aggregate: "SUM",
	}).Result()
	if err != nil {
		fmt.Printf("计算交集时出错: %v\n", err)
		return nil, err
	}

	err = u.client.ZAdd(ctx, likeRankKey, redis.Z{
		Score:  userLikeCount.Val(),
		Member: userID,
	}).Err()
	if err != nil {
		return nil, err
	}

	descResults, err := u.client.ZRevRangeWithScores(ctx, likeRankKey, 0, -1).Result()
	if err != nil {
		fmt.Printf("获取降序结果时出错: %v\n", err)
		return nil, err
	}

	r := 1
	var userLikeRanks []*user_model.UserLikeRank
	for _, result := range descResults {
		userID, err := strconv.Atoi(result.Member.(string))
		if err != nil {
			return nil, err
		}

		userLikeRank := &user_model.UserLikeRank{
			UserID: userID,
			Likes:  int(result.Score),
			Rank:   r,
		}
		r++
		userLikeRanks = append(userLikeRanks, userLikeRank)
	}
	return userLikeRanks, nil
}

func (u *UserCacheRepository) AddItem(item *user_model.CommunityItem) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:item:%d:info", UserCachePrefix, item.ID)
	if err := u.client.HSet(ctx, key, "id", item.ID, "name", item.Name, "price", item.Price, "capacity",
		item.Capacity, "remain", item.Remain, "begin", item.Begin).Err(); err != nil {
		return fmt.Errorf("UserCacheRepository.AddItem err:%w", err)
	}
	return nil
}

func (u *UserCacheRepository) GetAllItems() ([]*user_model.CommunityItem, error) {
	var items []*user_model.CommunityItem
	var cursor uint64
	ctx := context.Background()
	for {
		var keys []string
		var err error
		keys, nextCursor, err := u.client.Scan(ctx, cursor, "hcl:user:item:*:info", 0).Result()
		if err != nil {
			return nil, fmt.Errorf("UserCacheRepository.GetAllItems err:%w", err)
		}
		for _, key := range keys {
			itemMap, err := u.client.HGetAll(ctx, key).Result()
			if err != nil {
				return nil, fmt.Errorf("UserCacheRepository.GetAllItems err:%w", err)
			}
			id, err := strconv.Atoi(itemMap["id"])
			if err != nil {
				return nil, fmt.Errorf("UserCacheRepository.GetAllItems err:%w", err)
			}
			price, err := strconv.Atoi(itemMap["price"])
			if err != nil {
				return nil, fmt.Errorf("UserCacheRepository.GetAllItems err:%w", err)
			}
			capacity, err := strconv.Atoi(itemMap["capacity"])
			if err != nil {
				return nil, fmt.Errorf("UserCacheRepository.GetAllItems err:%w", err)
			}
			remain, err := strconv.Atoi(itemMap["remain"])
			if err != nil {
				return nil, fmt.Errorf("UserCacheRepository.GetAllItems err:%w", err)
			}
			layout := time.RFC3339Nano
			begin, err := time.Parse(layout, itemMap["begin"])
			if err != nil {
				return nil, fmt.Errorf("UserCacheRepository.GetAllItems err:%w", err)
			}
			items = append(items, &user_model.CommunityItem{
				ID:       id,
				Name:     itemMap["name"],
				Price:    price,
				Capacity: capacity,
				Remain:   remain,
				Begin:    begin,
			})
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return items, nil
}

func (u *UserCacheRepository) ChooseItem(userID int, itemID int) error {
	ctx := context.Background()

	// 定义Lua脚本
	script := `
    -- KEYS[1]: 用户集合键 
    -- KEYS[2]: 库存键 
    -- ARGV[1]: 用户ID
    
    -- 检查用户是否已在集合中
    local userExists = redis.call('SISMEMBER', KEYS[1], ARGV[1])
    if userExists == 1 then
        return 0  -- 用户已存在，返回0表示失败
    end
    
    -- 检查库存是否大于0
    local remain = tonumber(redis.call('HGET', KEYS[2],'remain'))
    if  remain==nil or remain<=0 then
        return remain  -- 库存不足，返回-1表示失败
    end
    
    -- 减少库存
	redis.call('HINCRBY', KEYS[2],'remain',-1)
    
    -- 将用户添加到集合中
    redis.call('SADD', KEYS[1], ARGV[1])
    
    return 1  -- 操作成功，返回1表示成功
    `

	// 构建键名
	usersKey := fmt.Sprintf("%s:item:%d:users", UserCachePrefix, itemID)
	itemKey := fmt.Sprintf("%s:item:%d:info", UserCachePrefix, itemID)

	// 执行Lua脚本
	result, err := u.client.Eval(ctx, script, []string{usersKey, itemKey}, userID).Result()
	if err != nil {
		return fmt.Errorf("执行Lua脚本失败: %w", err)
	}

	res, ok := result.(int64)
	if !ok {
		return fmt.Errorf("解析结果失败: %w", err)
	}
	// 处理返回结果
	switch res {
	case 1:
		return nil // 操作成功
	case 0:
		return fmt.Errorf("用户已选择此商品")
	case -1:
		return fmt.Errorf("库存不足")
	default:
		return fmt.Errorf("未知返回值: %d", result)
	}
}
