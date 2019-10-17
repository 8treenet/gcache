package internal

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/8treenet/gcache/option"

	"github.com/jinzhu/gorm"
)

const (
	modelKey          = "%s_model_%s"  //模型key
	searchKey         = "%s_search_%s" //查询主键列表 key
	affectKey         = "%s_affect_%s" //影响key
	checkTimeoutSec   = 3600
	skipCache         = "cache:skip_cache"          // 跳过缓存
	whereModelsSearch = "cache:where_models_search" //join和select查询
	whereIndex        = "cache:where_index"
)

func newHandleManager(db *gorm.DB, cp *plugin, redisOption *option.RedisOption) *Handle {
	result := new(Handle)
	result.redisClient = newRedisClient(redisOption)
	result.db = db
	result.cp = cp
	result.cleaner = make(map[string]*struct {
		unix   int64
		search bool
	})
	rand.Seed(time.Now().UnixNano())
	return result
}

type Handle struct {
	redisClient RedisClient
	db          *gorm.DB
	cp          *plugin
	cleaner     map[string]*struct {
		unix   int64
		search bool
	}
	cleanerMutex sync.Mutex
	debug        bool
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

func (h *Handle) JoinSearchKey(table string, key string, indexKeys []interface{}) string {
	result := fmt.Sprintf(searchKey, table, key)
	if len(indexKeys) == 0 {
		return result
	}

	var sk []string
	for _, key := range indexKeys {
		sk = append(sk, "tag:"+fmt.Sprint(key))
	}
	result = fmt.Sprintf("%s_%s", result, strings.Join(sk, "_")) + "_"
	return result
}

func (h *Handle) JoinAffectKey(table string, key string) string {
	return fmt.Sprintf(affectKey, table, key)
}

func (h *Handle) JoinCountSecondKey(key string) string {
	return "count:" + key
}

func (h *Handle) RefreshEvent(key string, search bool) {
	defer h.cleanerMutex.Unlock()
	h.cleanerMutex.Lock()
	_, ok := h.cleaner[key]
	if !ok {
		var item struct {
			unix   int64
			search bool
		}
		item.search = search
		item.unix = time.Now().Unix() + int64(checkTimeoutSec+rand.Intn(checkTimeoutSec))
		h.cleaner[key] = &item
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
		//每30秒检查一次 cleaner map 是否该处理
		time.Sleep(30 * time.Second)
		nowUnix := time.Now().Unix()
		h.cleanerMutex.Lock()
		for key, item := range h.cleaner {
			if item.unix+checkTimeoutSec > nowUnix {
				continue
			}

			dh := h.NewDeleteHandle()
			go dh.refresh(key, item.search)
			item.unix = nowUnix + int64(checkTimeoutSec+rand.Intn(checkTimeoutSec))
		}
		h.cleanerMutex.Unlock()
	}
}

type JsonModel struct {
	PK    string
	Model interface{}
}

type JsonSearch struct {
	Timeout  int64
	Primarys []interface{}
}
