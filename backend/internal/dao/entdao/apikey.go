package entdao

import (
	"context"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/ent"
	"github.com/Lestine-Yan/irisImg/backend/ent/apikey"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// apiKeyDAO 是 dao.APIKeyDAO 的 Ent 实现。
type apiKeyDAO struct {
	client *ent.Client
}

// NewAPIKeyDAO 基于 Ent 客户端构造 dao.APIKeyDAO。
func NewAPIKeyDAO(client *ent.Client) dao.APIKeyDAO {
	return &apiKeyDAO{client: client}
}

// 编译期断言：apiKeyDAO 必须实现 dao.APIKeyDAO。
var _ dao.APIKeyDAO = (*apiKeyDAO)(nil)

func (d *apiKeyDAO) Create(ctx context.Context, key *model.APIKey) (*model.APIKey, error) {
	row, err := d.client.ApiKey.Create().
		SetName(key.Name).
		SetKeyHash(key.KeyHash).
		SetPrefix(key.Prefix).
		SetScope(apikey.Scope(key.Scope)).
		SetRateLimit(key.RateLimit).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toAPIKeyModel(row), nil
}

func (d *apiKeyDAO) GetByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	row, err := d.client.ApiKey.Query().
		Where(apikey.KeyHashEQ(hash)).
		First(ctx)
	if err != nil {
		return nil, wrapErr(err)
	}
	return toAPIKeyModel(row), nil
}

func (d *apiKeyDAO) GetByID(ctx context.Context, id int) (*model.APIKey, error) {
	row, err := d.client.ApiKey.Get(ctx, id)
	if err != nil {
		return nil, wrapErr(err)
	}
	return toAPIKeyModel(row), nil
}

func (d *apiKeyDAO) List(ctx context.Context) ([]*model.APIKey, error) {
	rows, err := d.client.ApiKey.Query().
		Order(ent.Desc(apikey.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*model.APIKey, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAPIKeyModel(row))
	}
	return items, nil
}

func (d *apiKeyDAO) Revoke(ctx context.Context, id int) error {
	if err := d.client.ApiKey.UpdateOneID(id).
		SetRevoked(true).
		Exec(ctx); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (d *apiKeyDAO) TouchLastUsed(ctx context.Context, id int, t time.Time) error {
	if err := d.client.ApiKey.UpdateOneID(id).
		SetLastUsedAt(t).
		Exec(ctx); err != nil {
		return wrapErr(err)
	}
	return nil
}

func (d *apiKeyDAO) UpdateName(ctx context.Context, id int, name string) (*model.APIKey, error) {
	row, err := d.client.ApiKey.UpdateOneID(id).
		SetName(name).
		Save(ctx)
	if err != nil {
		return nil, wrapErr(err)
	}
	return toAPIKeyModel(row), nil
}

// ResetKey 写入新的哈希与前缀，并清除吊销状态，使该密钥重新可用。
func (d *apiKeyDAO) ResetKey(ctx context.Context, id int, keyHash, prefix string) (*model.APIKey, error) {
	row, err := d.client.ApiKey.UpdateOneID(id).
		SetKeyHash(keyHash).
		SetPrefix(prefix).
		SetRevoked(false).
		Save(ctx)
	if err != nil {
		return nil, wrapErr(err)
	}
	return toAPIKeyModel(row), nil
}

func (d *apiKeyDAO) Delete(ctx context.Context, id int) error {
	if err := d.client.ApiKey.DeleteOneID(id).Exec(ctx); err != nil {
		return wrapErr(err)
	}
	return nil
}

// toAPIKeyModel 将 Ent 实体转换为跨层的 model.APIKey。
func toAPIKeyModel(e *ent.ApiKey) *model.APIKey {
	if e == nil {
		return nil
	}
	return &model.APIKey{
		ID:         e.ID,
		Name:       e.Name,
		Prefix:     e.Prefix,
		Scope:      string(e.Scope),
		KeyHash:    e.KeyHash,
		RateLimit:  e.RateLimit,
		Revoked:    e.Revoked,
		LastUsedAt: e.LastUsedAt,
		CreatedAt:  e.CreatedAt,
	}
}
