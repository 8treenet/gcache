# DHT
##### 一致性哈希

###### 创建和使用
```go
import (
	"github.com/8treenet/venus/dht"
)
func TestNormal(t *testing.T) {
	//创建一致性哈希和范围1-100的节点
	hash := dht.New().Range(1, 100)
	t.Log(hash.FindNode(55)) //查找节点 55

	//创建一致性哈希和列表节点
	hash = dht.New().List("hostname1", "hostname2", "hostname3")
	t.Log(hash.FindNode("hostname1")) //查找节点 hostname1

	//获取全部节点
	for _, v := range hash.GetNodes() {
		//打印节点
		t.Log(v.Value(), v.CRC32())
	}
}
```

```sh
$ 55
$ hostname1
$ hostname1 3918110341
$ hostname2 1887489855
$ hostname3 126353321
```


###### 查找数据的节点
```go
import (
	"github.com/8treenet/venus/dht"
)
func TestSearch(t *testing.T) {
	//方式1 创建一致性哈希和范围1-100的节点
	hash := dht.New().Range(1, 100)
	//输入数据 查找节点
	node := hash.Search("freedom")
	t.Log(node.Value(), node.CRC32())

	//方式2 创建一致性哈希和列表节点
	hash = dht.New().List("hostname1", "hostname2", "hostname3")
	//输入数据 查找节点
	node = hash.Search("group-1001")
	t.Log(node.Value(), node.CRC32())
}
```
```sh
$ 74 4033496702
$ hostname3 126353321
```

###### 节点伸缩后的数据分布
```go
import (
	"fmt"
	"math/rand"
	"testing"
	"github.com/8treenet/venus/dht"
)
func TestRebalance(t *testing.T) {
	//创建一致性哈希和范围1-100的节点
	hash := dht.New().Range(1, 100)

	//伪造50000行数据
	rows := []string{}
	for i := 0; i < 50000; i++ {
		rows = append(rows, fmt.Sprintf("freedom:%d", rand.Intn(99999999999)))
	}

	//为每行数据分配crc32
	rowNodeMap := map[string]uint32{}
	for i := 0; i < len(rows); i++ {
		node := hash.Search(rows[i]) //查找节点
		rowNodeMap[rows[i]] = node.CRC32()
	}

	//测试伸缩节点后crc32的分布
	diffCount := 0
	hash.AddNode(dht.NewNode(101)) //增加节点101
	//hash.RemoveNode(dht.NewNode(88)) //删除节点88
	for i := 0; i < len(rows); i++ {
		node := hash.Search(rows[i]) //重新查找节点
		if rowNodeMap[rows[i]] != node.CRC32() {
			diffCount++ //重新分布总数递增
		}
	}

	t.Log(diffCount) //打印重新分布的总数
}
```
```sh
$ 1374
```