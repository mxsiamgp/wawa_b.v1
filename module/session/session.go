package session

import (
	"time"
	"strings"

	"github.com/labstack/echo"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/satori/go.uuid"
	"gopkg.in/redis.v4"
)

const CTX_KEY_SESSION = "SESSION.SESSION"

// 会话存储器
type SessionStore interface {
	// 获取会话的所有键值对
	Get(sessID string) (kvsJSON string, ok bool)

	// 设置会话的所有键值对
	Set(sessID string, kvsJSON string)
}

// Redis会话存储器
type RedisSessionStore struct {
	// Redis客户端
	client     *redis.Client

	// 过期时间
	expiration time.Duration
}

// 会话
type Session struct {
	// 缓存
	cache map[string]string

	// ID
	id    string

	// 存储器
	store SessionStore
}

type SessionManager struct {
	// Cookie键
	cookieName string

	// 存储器
	store      SessionStore
}

// 创建一个Redis会话存储器
func NewRedisSessionStore(client *redis.Client, expiration time.Duration) *RedisSessionStore {
	return &RedisSessionStore{
		client: client,
		expiration: expiration,
	}
}

// 获取会话的所有键值对
func (store *RedisSessionStore) Get(sessID string) (string, bool) {
	existsStat := store.client.Exists(sessID)
	if existsStat.Err() != nil {
		panic(existsStat.Err())
	}

	if !existsStat.Val() {
		return "", false
	}

	getStat := store.client.Get(sessID)

	return getStat.Val(), true
}

// 设置会话的所有键值对
func (store *RedisSessionStore) Set(sessID string, kvsJSON string) {
	if stat := store.client.Set(sessID, kvsJSON, store.expiration); stat.Err() != nil {
		panic(stat.Err())
	}
}

// 创建一个会话
func NewSession(id string, store SessionStore) *Session {
	return &Session{
		id: id,
		cache: make(map[string]string),
		store: store,
	}
}

// 获取键对应的值
func (sess *Session) Get(key string, val interface{}) bool {
	valJSON, ok := sess.cache[key]
	if !ok {
		return false
	}
	if err := ffjson.Unmarshal([]byte(valJSON), val); err != nil {
		panic(err)
	}
	return true
}

// 设置键对应的值
func (sess *Session) Set(key string, val interface{}) {
	valJSON, err := ffjson.Marshal(val)
	if err != nil {
		panic(err)
	}
	sess.cache[key] = string(valJSON)
}

// 删除键对应的值
func (sess *Session) Remove(key string) {
	delete(sess.cache, key)
}

// 从存储器重新载入键值对
func (sess *Session) Load() bool {
	kvsJSON, ok := sess.store.Get(sess.id)
	if ok {
		kvs := make(map[string]string)
		err := ffjson.Unmarshal([]byte(kvsJSON), &kvs)
		if err != nil {
			panic(err)
		}
		sess.cache = kvs
	}
	return ok
}

// 保存键值对到存储器
func (sess *Session) Save() {
	kvsJSON, err := ffjson.Marshal(sess.cache)
	if err != nil {
		panic(err)
	}
	sess.store.Set(sess.id, string(kvsJSON))
}

// 创建一个会话管理器
func NewManager(cookieName string, store SessionStore) *SessionManager {
	return &SessionManager{
		cookieName: cookieName,
		store: store,
	}
}

// 获取Echo中间件
func (mgr *SessionManager) HandlerFunc() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			var sid string
			for _, ck := range ctx.Cookies() {
				if ck.Name() == mgr.cookieName {
					sid = ck.Value()
					break
				}
			}

			if sid == "" {
				sid = strings.Replace(uuid.NewV4().String(), "-", "", -1)
				ck := new(echo.Cookie)
				ck.SetName(mgr.cookieName)
				ck.SetValue(sid)
				ck.SetHTTPOnly(true)
				ctx.SetCookie(ck)
			}

			sess := NewSession(sid, mgr.store)
			sess.Load()
			ctx.Set(CTX_KEY_SESSION, sess)
			defer sess.Save()

			return next(ctx)
		}
	}
}

// 获取Echo上下文关联的会话
func GetSessionByContext(ctx echo.Context) *Session {
	return ctx.Get(CTX_KEY_SESSION).(*Session)
}
