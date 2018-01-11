# cache
使用方法：

//首先取一个cache的struct

//可以利用不同的patch进行批量删除

key:="luyu6056"

patch:="member_info"

cache := libraries.Hget(key,patch)

//读取一个缓存数据

last_login_time := cache.Load("login_time")

//储存一个临时缓存，重启进程失效，不会进行持久化写入

//cache.Stroe(key interface{},value interface{})

cache.Stroe("login_time",libraries.Timestamp())

//传入一个持久化数据，写入硬盘

member_info:=map[string]interface{}{"age":18,"sex":"man","birthday":"1970-01-01"}

//Hset方法，默认每秒钟写入一次硬盘

libraries.Hset(key,member_info,patch,0)

//Hset_r方法，立即写入硬盘

libraries.Hset_r(key,member_info,patch,86400)//整个缓存有效时间86400，超时该key清空所有数据

//删除一条key

libraries.Hdel(key,patch)

//删除整个patch

libraries.Hdel_all(patch)

# 队列
以redis的list为参照集成了RPUSH、LPUSH、LPOP、RPOP、LINDEX、LLEN、LRANGE、LREM、LTRIM
func pop_test() {
	begin := Timestampint() + 1
	fmt.Println("开始测试")
	//读取左边数据等待100秒
	//线程1
	go func() {
		fmt.Println(LPOP("test", 100))
		fmt.Println("1等待了", Timestampint()-begin, "秒")
	}()
	//延迟1秒执行线程2
	time.Sleep(time.Second * 1)
	//线程2
	go func() {
		fmt.Println(LPOP("test", 10))
		fmt.Println("2等待了", Timestampint()-begin, "秒")
	}()
	//等5秒后再写入
	time.Sleep(time.Second * 5)
	fmt.Println("开始写入1")
	RPUSH("test", "久等了")
	//等待3秒后写入
	time.Sleep(time.Second * 3)
	fmt.Println("开始写入2")
	LPUSH("test", "第二次写入")
}

func lrange_test() {
	LPUSH("test", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
	fmt.Println(LRANGE("test", 0, 1))
	fmt.Println(LRANGE("test", 5, 10))
	fmt.Println(LRANGE("test", -2))
}

func lrem_test() {
	LPUSH("test", "5", "2", "2", "3", "3", "3", "4", "5", "6", "7")
	LREM("test", 0, "2")          //去掉所有的2
	fmt.Println(list_map["test"]) //[5 3 3 3 4 5 6 7]
	LREM("test", 2, "3")          //去掉左边两个3
	fmt.Println(list_map["test"]) //[5 3 4 5 6 7]
	LREM("test", -1, "5")         //去掉右边那个5
	fmt.Println(list_map["test"]) //[5 3 4   6 7]
}

func ltrim_test() {
	LPUSH("test", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
	fmt.Println(LTRIM("test", 0, 7))
	fmt.Println(list_map["test"]) //[0 1 2 3 4 5 6 7]
	fmt.Println(LTRIM("test", 2, 4))
	fmt.Println(list_map["test"]) //[2 3 4]
	fmt.Println(LTRIM("test", 10))
	fmt.Println(list_map["test"]) //[2 3 4]
}
