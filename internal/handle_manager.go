package internal

import (
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

const (
	modelKey        = "%s_model_%s"  //模型key
	searchKey       = "%s_search_%s" //查询主键列表 key
	affectKey       = "%s_affect_%s" //影响key
	checkTimeoutSec = 1800
	dontInvalidSearch = "cache:dont_update_search"  //强制不失效
	whereModelsSearch = "cache:where_models_search" //join和select查询
)


func newHandleManager(db *gorm.DB, cp *plugin) *Handle {
	result := new(Handle)
	rclient := redis.NewClient(&redis.Options{
		Addr:     cp.defaultOpt.RedisAddr,
		Password: cp.defaultOpt.RedisPassword,
		DB:       cp.defaultOpt.RedisDB,
		OnConnect: func(conn *redis.Conn) error {
			return nil
		},

		//MaxRetries: 0,                       默认不重试
		//PoolSize:  5, 					   连接池最大连接数 默认cpu * 10
		//ReadTimeout : time.Duration 	       默认3秒
		//WriteTimeout: time.Duration          默认3秒
		//MinIdleConns: 0				       最小常驻空闲连接
		//MaxConnAge : time.Duration           连接最长时间，默认永久
		//IdleTimeout: time.Duration		   连接空闲时间 默认5分钟
		//IdleCheckFrequency: time.Duration    保活，检测连接 默认1分钟
		//PoolTimeout time.Duration            如果连接池已满 等待可用连接的时间默认 4
	})
	result.redisClient = rclient
	result.db = db
	result.cp = cp
	result.cleaner = make(map[reflect.Type]int64)
	if perr := rclient.Ping().Err(); perr != nil {
		panic(perr)
	}
	rand.Seed(time.Now().UnixNano())
	return result
}

type Handle struct {
	redisClient *redis.Client
	db          *gorm.DB
	cp          *plugin
	cleaner     map[reflect.Type]int64
	cleanerMutex     sync.Mutex
	debug       bool
}

func (h *Handle) NewDeleteHandle() *deleteHandle {
	return newDeleteHandle(h)
}
func (h *Handle) NewCreateHandle() *createHandle {
	return newCreateHandle(h)
}

func (h *Handle) NewQueryHandle() *queryHandle {
	return newQueryHandle(h)
}

func (h *Handle) NewUpdateHandle() *updateHandle {
	return newUpdateHandle(h)
}

func (h *Handle) registerCall() {
	newCallCreate(h).Bind()
	newCallDelete(h).Bind()
	newCallQuery(h).Bind()
	newCallRow(h).Bind()
	newCallUpdate(h).Bind()
}

func (h *Handle) JoinModelKeys(table string, primarys ...interface{}) []string {
	var keys []string
	for index := 0; index < len(primarys); index++ {
		key := fmt.Sprintf(modelKey, table, fmt.Sprint(primarys[index]))
		keys = append(keys, key)
	}
	return keys
}

func (h *Handle) JoinModelKey(table string, primary interface{}) string {
	return fmt.Sprintf(modelKey, table, fmt.Sprint(primary))
}

func (h *Handle) JoinSearchKey(table string, key string) string {
	return fmt.Sprintf(searchKey, table, key)
}

func (h *Handle) JoinAffectKey(table string, key string) string {
	return fmt.Sprintf(affectKey, table, key)
}

func (h *Handle) JoinCountSecondKey(key string) string {
	return "count:" + key
}

func (h *Handle) RefreshEvent(t reflect.Type)  {
	defer h.cleanerMutex.Unlock()
	h.cleanerMutex.Lock()

	_, ok := h.cleaner[t]
	if !ok {
		h.cleaner[t] = time.Now().Unix() + checkTimeoutSec + rand.Int63n((int64(checkTimeoutSec / 2)))
	}
	return
}

func (h *Handle) Debug(item ...interface{}) {
	if !h.debug {
		return
	}
	fmt.Println(item...)
}

func (h *Handle) RefreshRun() {
	for {
		//每10秒检查一次 cleaner map 是否该处理
		time.Sleep(10 * time.Second)
		nowUnix := time.Now().Unix()
		h.cleanerMutex.Lock()
		for table, updateUnix := range h.cleaner {
			if updateUnix+checkTimeoutSec > nowUnix {
				continue
			}

			// table
			dh := h.NewDeleteHandle()
			go dh.refresh(table)
			h.cleaner[table] = nowUnix + checkTimeoutSec + rand.Int63n((int64(checkTimeoutSec / 2)))
		}
		h.cleanerMutex.Unlock()
	}
}

type JsonModel struct {
	PK    string
	Model interface{}
}

type JsonSearch struct {
	UpdatedAt int64
	Primarys  []interface{}
}
