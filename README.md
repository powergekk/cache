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
以redis的list为参照集成了RPUSH、LPUSH、LPOP、RPOP、LINDEX、LLEN、LRANGE、LREM、LTRIM，例子在cache.go尾部
