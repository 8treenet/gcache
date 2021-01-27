package example_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/TIZX/gcache"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	db          *gorm.DB
	cachePlugin gcache.Plugin
)

type TestUser struct {
	gorm.Model
	UserName string `gorm:"size:32"`
	Password string `gorm:"size:32"`
	Age      int
	Status   int
}

type TestEmail struct {
	gorm.Model
	TypeID     int
	Subscribed bool
	TestUserID int
}

func init() {
	var e error
	addr := "root:123123@tcp(127.0.0.1:3306)/freedom?charset=utf8&parseTime=True&loc=Local"
	db, e = gorm.Open("mysql", addr)
	if e != nil {
		panic(e)
	}
	db.AutoMigrate(&TestUser{})
	db.AutoMigrate(&TestEmail{})

	opt := gcache.DefaultOption{}
	opt.Expires = 300              //缓存时间，默认60秒。范围 30-900
	opt.Level = gcache.LevelSearch //缓存级别，默认LevelSearch。LevelDisable:关闭缓存，LevelModel:模型缓存， LevelSearch:查询缓存
	opt.AsyncWrite = false         //异步缓存更新, 默认false。 insert update delete 成功后是否异步更新缓存
	opt.PenetrationSafe = false    //开启防穿透, 默认false。

	//缓存中间件 注入到Gorm
	cachePlugin = gcache.AttachDB(db, &opt, &gcache.RedisOption{Addr: "localhost:6379"})

	InitData()
	//开启Debug，查看日志
	db.LogMode(true)
	cachePlugin.Debug()
}

func InitData() {
	cachePlugin.FlushDB()
	db.Exec("truncate test_users")
	db.Exec("truncate test_emails")
	for index := 1; index < 21; index++ {
		user := &TestUser{}
		user.UserName = fmt.Sprintf("%s_%d", "name", index)
		user.Password = fmt.Sprintf("%s_%d", "password", index)
		user.Age = 20 + index
		user.Status = rand.Intn(3)
		db.Save(user)

		email := &TestEmail{}
		email.TypeID = index
		email.TestUserID = index
		db.Save(email)
	}
}

func Two(fun func()) {
	fmt.Println("one:")
	fun()
	fmt.Print("\n\n\n\n\n")
	fmt.Println("two:")
	fun()
}

/*
	简单主键模型查询
	同条件查询，第一次未命中，第二次命中
	ps: Two(func()) : 执行2次
*/
func TestQuerySimple(t *testing.T) {
	Two(func() {
		var tcs []*TestUser
		var tcs2 []TestUser
		db.Find(&tcs, []int{1, 2})
		fmt.Println("1", tcs)
		db.Find(&tcs2, []int{1, 2})
		fmt.Println("2", tcs2)

		var tc TestUser
		db.First(&tc, 1)
		fmt.Println("3", tc)
		db.Last(&tc, 1)
		fmt.Println("4", tc)
		db.Where("id = ?", 1, &tc)
		fmt.Println("5", tc)
	})
}

/*
	模型关联
*/
func TestQueryRelated(t *testing.T) {
	Two(func() {
		var tc TestUser
		var ems []TestEmail
		db.First(&tc, 1).Related(&ems)
		fmt.Println(tc, ems)
	})
}

/*
	条件查询
	1. where条件使用?, 查询条件和查询参数必须严格区分。
*/
func TestQueryWhere(t *testing.T) {
	Two(func() {
		var tc TestUser
		var tcs []TestUser
		var tcs2 []*TestUser
		var count int
		//错误的方式 db.Where("user_name = name_1").First(&tc)
		db.Where("user_name = ?", "name_1").First(&tc)
		fmt.Println("where 1", tc)
		tc = TestUser{}

		db.Where("user_name LIKE ?", "%_2%").Order("age desc").Find(&tcs)
		fmt.Println("where 2", tcs)
		tcs = []TestUser{}

		db.Where("password = ? and age = ?", "password_2", 22).First(&tc)
		fmt.Println("where 3", tc)
		tc = TestUser{}

		db.Where(&TestUser{UserName: "name_14", Age: 34}).First(&tc)
		fmt.Println("where 4", tc)
		tc = TestUser{}

		db.Where("user_name in (?)", []string{"name_17", "name_18", "name_19", "name_20"}).Limit(3).Find(&tcs).Count(&count)
		fmt.Println("where 5", tcs, count)
		tcs = []TestUser{}

		db.Where(map[string]interface{}{"user_name": "name_20", "password": "password_20"}).Find(&tcs).Count(&count)
		fmt.Println("where 6", tcs, count)
		tcs = []TestUser{}

		db.Find(&tcs, "user_name = ?", "name_11")
		fmt.Println("where 7", tcs)
		tcs = []TestUser{}

		db.Find(&tcs, &TestUser{UserName: "name_11"})
		fmt.Println("where 8", tcs)
		tcs = []TestUser{}

		db.First(&tc, "user_name =? and age = ?", "name_7", 27)
		fmt.Println("where 9", tc)
		tc = TestUser{}

		db.Find(&tcs2, &TestUser{UserName: "name_11"})
		fmt.Println("where 10", tcs2)
		tcs2 = []*TestUser{}

		db.Where("user_name in (?)", []string{"name_17", "name_18", "name_19", "name_20"}).Limit(3).Find(&tcs2).Count(&count)
		fmt.Println("where 11", tcs2, count)
		tcs2 = []*TestUser{}
	})
}

func TestQueryNot(t *testing.T) {
	Two(func() {
		var tc TestUser
		var tcs []TestUser
		var count int
		db.Not("user_name", "not_1_yangshu").Order("id,age desc").Limit(5).Find(&tcs).Count(&count)
		db.Not("user_name = ?", "not_2_yangshu").First(&tc)

		fmt.Println(tcs, count)
		fmt.Println(tc)
	})
}

func TestQueryOr(t *testing.T) {
	Two(func() {
		var tcs []TestUser
		var count int
		db.Where("age = ?", 21).Or("age = ?", 30).Find(&tcs).Count(&count)
		fmt.Println(tcs, count)
	})
}

func TestQueryOrder(t *testing.T) {
	Two(func() {
		var tcs []TestUser
		var count int
		var ages []int64
		db.Where("user_name = ?", "Yangshu").Order("id desc").Offset(0).Limit(10).Find(&tcs).Count(&count)
		fmt.Println(tcs, count, ages)
	})
}

/*
	join查询缓存
	cachePlugin.CreateRelative(model ...interface{}) : 传入要join的模型，辅助缓存做关联。
*/
func TestQueryJoin(t *testing.T) {
	Two(func() {
		var tcs []TestUser
		var count int
		cachePlugin.CreateRelative(&TestEmail{}).Joins("left join test_emails on test_emails.test_user_id = test_users.id").Where("type_id > ?", 18).Find(&tcs).Count(&count)
		fmt.Println(tcs, count)
	})
}

/*
	子查询缓存
	cachePlugin.CreateRelative(model ...interface{}) : 传入要子查询的模型，辅助缓存做关联。
*/
func TestSelect(t *testing.T) {
	Two(func() {
		var count int
		var tcs []TestUser
		cachePlugin.CreateRelative(&TestEmail{}).Where("id in(select test_user_id from test_emails where type_id > ?)", 18).Find(&tcs).Count(&count)
		fmt.Println(tcs, count)
	})
}

/*
	单独使用模型缓存配置
	修改 opt.Level = gcache.LevelDisable 查看输出
*/
func TestModelOpt(t *testing.T) {
	Two(func() {
		var te TestEmail
		db.Where("type_id >= ?", 18).First(&te)
		fmt.Println(te)
	})
}

//Cache 重写模型单独配置项
func (te *TestEmail) Cache(opt *gcache.ModelOption) {
	opt.Expires = 600
	opt.Level = gcache.LevelSearch
	opt.AsyncWrite = false
	opt.PenetrationSafe = false
}

/*
	update后缓 缓存失效
*/
func TestUpdateInvalid1(t *testing.T) {
	var tcs []TestUser
	Two(func() {
		db.Where("age = ?", 39).Or("age = ?", 40).Find(&tcs)
		fmt.Println("Query 1", tcs)
	})

	for index := 0; index < len(tcs); index++ {
		//触发缓存失效
		//ps : 只失效 age 字段影响的查询缓存
		db.Model(&tcs[index]).Update("age", tcs[index].Age+1)
	}

	tcs = []TestUser{}
	Two(func() {
		var count int
		db.Where("age = ?", 40).Or("age = ?", 41).Find(&tcs).Count(&count)
		fmt.Println("Query 2", tcs, count)
	})

	for index := 0; index < len(tcs); index++ {
		//触发缓存失效
		db.Model(&tcs[index]).Updates(map[string]interface{}{"age": tcs[index].Age - 1, "password": "1111"})
	}
}

/*
	TestUpdateInvalid2 join失效
*/
func TestUpdateInvalid2(t *testing.T) {
	//join查询 填充缓存
	Two(func() {
		var tcs []TestUser
		var count int
		cachePlugin.CreateRelative(&TestEmail{}).Joins("left join test_emails on test_emails.test_user_id = test_users.id").Where("type_id >= ?", 18).Find(&tcs).Count(&count)
		fmt.Println(tcs, count)
	})

	//update失效
	var te TestEmail
	//join查询使用了 test_emails.type_id 字段，更新 test_emails.type_id 会触发join的失效
	db.Where("type_id >= ?", 18).First(&te).Update("type_id", te.TypeID+30)
}

/*
	TestUpdateInvalid3
	非主键update条件，删除相关缓存模型
*/
func TestUpdateInvalid3(t *testing.T) {
	//主键查询 填充缓存
	var tcs []TestUser
	db.Find(&tcs, []int{1, 2, 3, 4, 5, 6, 7, 8, 9})

	/*
		触发缓存失效,内部执行顺序。
		1. 先执行 select id from test_users where (update条件) => idList
		2. 执行 sql update
		3. 删除缓存 redis delete idList
	*/
	db.Model(&TestUser{}).Where("age < ?", 30).Update("age", 100)

	/*
		如果条件数据较多，建议外部循环update
		for {
			db.Model(&TestUser{}).Where("age < ?", 30).Limit(20).Update("age", 100)
		}
	*/
}

/*
	TestDeleteInvalid
	删除数据，使缓存失效
*/
func TestDeleteInvalid(t *testing.T) {
	//join查询 填充缓存
	var tcs []TestUser
	var count int
	cachePlugin.CreateRelative(&TestEmail{}).Joins("left join test_emails on test_emails.test_user_id = test_users.id").Where("type_id >= ?", 18).Find(&tcs).Count(&count)

	//普通查询 填充缓存
	var te TestEmail
	db.Where("type_id >= ?", 18).First(&te)

	//触发缓存失效
	db.Delete(&te)
}

/*
	TestCreateInvalid
	新增数据，使缓存失效
*/
func TestCreateInvalid(t *testing.T) {
	//join查询 填充缓存
	var tcs []TestUser
	var count int
	cachePlugin.CreateRelative(&TestEmail{}).Joins("left join test_emails on test_emails.test_user_id = test_users.id").Where("type_id >= ?", 18).Find(&tcs).Count(&count)

	//普通查询 填充缓存
	var te TestEmail
	db.Where("type_id >= ?", 18).First(&te)

	//新增数据 触发缓存失效
	email := &TestEmail{}
	email.TypeID = 1101
	email.TestUserID = 1234
	db.Save(email)
}

/*
	Tag测试
	cachePlugin.CreateTag(...inteface{}) : 传入标签数据
*/
func TestTag(t *testing.T) {
	var tcs1 []TestUser
	var count1 int
	//填充tag `1` 的缓存
	cachePlugin.SetTag(db.Where("status = ?", 1), 1, 345).Find(&tcs1).Count(&count1)

	fmt.Println("tcs1", tcs1, count1)

	var tcs2 []TestUser
	var count2 int
	//填充tag `2` 的缓存
	cachePlugin.CreateTag(2).Where("status = ?", 2).Find(&tcs2).Count(&count2)
	fmt.Println("tcs2", tcs2, count2)
	if len(tcs2) == 0 {
		return
	}

	//update 使tag`2`失效
	fmt.Println("update", cachePlugin.CreateTag(2).Model(&tcs2[0]).Update("status", 6).RowsAffected)
	//delete 使tag`2`失效
	fmt.Println("delete", cachePlugin.CreateTag(2).Delete(&tcs2[0]).RowsAffected)

	//save 使tag`2`失效
	tcs2[0].ID = 0
	fmt.Println("save", cachePlugin.CreateTag(2).Save(&tcs2[0]).RowsAffected)
}

/*
	防击穿测试
*/
func TestSingleFlight(t *testing.T) {
	wait := new(sync.WaitGroup)
	for index := 0; index < 200; index++ {
		go func() {
			wait.Add(1)
			var tcs []TestUser
			db.Where([]int{1, 2}).Find(&tcs)
			fmt.Println(tcs)

			var count int
			tcs = []TestUser{}
			cachePlugin.CreateRelative(&TestEmail{}).Joins("left join test_emails on test_emails.test_user_id = test_users.id").Where("type_id >= ?", 18).Find(&tcs).Count(&count)
			fmt.Println(tcs, count)
			wait.Done()
		}()
	}
	wait.Wait()
}

/*
	防穿透测试, 可注释和解注观察效果。
	开启防穿透:查询数据库不存在的数据也会录入缓存
*/
func TestPenetration(t *testing.T) {
	Two(func() {
		var tcs []TestUser
		var count int
		db.Find(&tcs, []int{100, 200})
		fmt.Println(tcs)

		tcs = []TestUser{}
		db.Where("user_name = ?", "不存在").Find(&tcs).Count(&count)
		fmt.Println(tcs, count)
	})
}

//Cache 重写模型单独配置项
func (te *TestUser) Cache(opt *gcache.ModelOption) {
	//解注开启防穿透
	//opt.PenetrationSafe = true
}

/*
	手动删除
*/
func TestPlugin(t *testing.T) {
	//通过主键 删除模型缓存
	cachePlugin.DeleteModel(&TestEmail{}, 1)

	//删除模型的所有查询缓存
	cachePlugin.DeleteSearch(&TestEmail{})

	//删库
	cachePlugin.FlushDB()
}

/*
	增删查改 跳过缓存处理
*/
func TestSkipCache(t *testing.T) {
	var tc TestUser
	var ems []TestEmail
	cachePlugin.SkipCache().First(&tc, 1).Related(&ems)

	cachePlugin.SkipCache().Model(&tc).Update("age", tc.Age+1)

	cachePlugin.SkipCache().Delete(&tc)

	email := &TestEmail{}
	email.TypeID = 2101
	email.TestUserID = 2234
	cachePlugin.SkipCache().Save(email)
}
