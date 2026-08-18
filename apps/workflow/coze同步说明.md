## coze同步说明

### 合并执行

分支：sync/upstream 

```shell
git remote add upstream https://github.com/coze-dev/coze-studio.git
git fetch upstream

git rebase --rebase-merges upstream/main
```



### 修正项目路径映射

> resources/conf  →  conf

```shell
cd backend
```

#### mac

```shell
# symlink
ln -s ../conf resources/conf

# 验证
ls -l resources
```

#### windows

```powershell
mklink /D resources\conf conf

# 验证
dir resources
```



### 模型管理

https://github.com/coze-dev/coze-studio/wiki/3.-%E6%A8%A1%E5%9E%8B%E9%85%8D%E7%BD%AE

官方源新增model_instance表

并根据，

```sql
select * from kv_entries where namespace = 'kv_model_ns' and key_data = 'do_not_use_old_model_key' limit 1
```

判断是否启用最新数据库维护方式使用模型信息



**社区版逻辑为通过http://localhost:8888/admin/#model-management维护模型管理。**

> 首次添加会增加do_not_use_old_model_key的记录



**坑：**参考社区版部署脚本，依赖backend/resources/static路径，鉴权跳转路由（未登录跳转/sign，依赖前端编译后产物），以及管理员名单

```go
staticFile := path.Join(cwd, "resources/static/index.html")

r.Static("/static", path.Join(cwd, "/resources/static"))
r.StaticFile("/favicon.png", "./resources/static/favicon.png")
r.StaticFile("/", staticFile)
r.StaticFile("/sign", staticFile)
```



do_not_use_old_model_key不存在情况，会通过读取.env，作为

```bash
# Settings for Model
# Model for agent & workflow
# add suffix number to add different models
export MODEL_PROTOCOL_0="ark"       # protocol
export MODEL_OPENCOZE_ID_0="100001" # id for record
export MODEL_NAME_0=""              # model name for show
export MODEL_ID_0=""                # model name for connection
export MODEL_API_KEY_0=""           # model api key
export MODEL_BASE_URL_0=""           # model base url
```

以及model yml配置，根据id，写入model_instance。

作为oldModels。



### 插件管理

```go
basePath := path.Join(cwd, "resources", "conf", "plugin")

err = loadPluginProductMeta(ctx, basePath)
```

插件信息来源：

- 环境配置中的插件 resources/conf/plugin
- 数据库中插件信息 

ListPluginProducts 函数优先使用环境配置
