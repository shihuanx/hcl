项目包含用户模块和文章模块

用户模块主要包括：登录注册、管理员认证、权限校验、关注、取关、点赞、获取好友点赞排行榜、秒杀社区物品相关。

文章模块主要包括：添加文章、2种文章显示结构（基本与全部）、文章点赞、更新以及缓存删除策略。

具体：
登录注册：使用双token实现用户无感刷新且提高安全性

权限校验：使用gin框架设计中间件进行权限校验

关注相关：使用redis无序集合实现关注功能，为每个用户维护关注集合和粉丝集合从而得到共同好友、猜你喜欢

点赞相关：使用redis有序集合实现点赞排行榜，使用redis无序集合防止用户重复点赞

秒杀相关：使用redis与lua脚本对物品容量进行预扣防止超卖，并且把用户加入集合防止重复购买，再通过消息队列转发给消费者实现流量削峰，并开启mysql事务确保原子操作，可以设置消费者的开启和结束时间，以及开启多个消费者

添加文章：文章结构包含基本结构（比如首页看到的所有文章），以及具体结构（点进文章显示全部信息），添加文章时像reids添加基本文章结构，使用mysql事务确保一致性，异步添加完整结构
查询文章：查询完整结构时会先从缓存查，缓存没有的话就从mysql查再异步写入缓存

文章点赞：为每一个文章维护一个点赞用户集合防止重复点赞，点赞量不及时同步到mysql而是定时回写来提高性能，因此在查文章完整结构时如果缓存不存在，从mysql获取除点赞外的字段，从redis获取文章基本结构中的点赞字段

更新文章：更新文章时采取删除缓存-更新数据库-删除缓存策略，防止缓存和数据库的数据不一致


