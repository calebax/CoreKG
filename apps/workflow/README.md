# coze 踩坑日记

## 路径映射

> resources/conf  →  conf

```shell
cd backend
```

### mac

```shell
# symlink
ln -s ../conf resources/conf

# 验证
ls -l resources
```

### windows

```powershell
mklink /D resources\conf conf

# 验证
dir resources
```



## 启动命令

```sh
export APP_ENV=49
# back目录下
go run main.go
```
根目录下.vscoed配置
env文件根据实际情况修改
```json
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, 1visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "coze backend",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/backend/main.go",
            "env": {
                "APP_ENV": "49"
            }
        }
    ]
}
```

## function call

- **embedding配置**

  前期知识库解析文件保存是因为我们没有配置embedding模型，但是在coze启动时没有校验，在运行解析文件时才会报错。页面上能看到报错信息时关于not found embedding 的信息

- **OCR配置**

  类似embedding，在coze官方知识库解析文件选择ocr图片时会报错
  
- **模型配置**

  如果要用agent嵌入工作流等插件function_call: true一定要打开，但是coze并没有做模型能力校验。如果没有打开就会直接在网页上报错。

  所有需要调用function call的模型不能选择带有深度思考的模型，否则会报错。

## 资源搜索

- **分词器**

  与线上环境分词器有所不同搜索结果有差异，开源版本搜索是对输入的单个倒排索引进行左右模糊搜索，线上环境使用的是match查询`{"wildcard":{"name":{"case_insensitive":true,"value":"*test测试double*"}`

  **早期只能单字搜索是因为es分词器部署问题，没有应用到分词器，es自动按照字符进行拆分**

  ```
  dsl
  
  {"query":{"bool":{"must":[{"term":{"space_id":{"value":"7560981806409859072"}}},{"wildcard":{"name":{"case_insensitive":true,"value":"*test测试double*"}}},{"terms":{"status":[1,3,4]}},{"terms":{"type":[1,2]}}]}},"size":25,"sort":[{"update_time":{"order":"desc"}}]}
  ```

## 开发调试

- **调试**

  因为coze在问答部分开发时使用了很多channel，所有不好去调试具体操作，并且不只一个channel所有在断点调试时会有线程错乱的情况，很可能调试一半时程序崩溃。



## 代码风格

代码风格较为统一，有严格的设计模式，值得学习。



## 工具分享

https://deepwiki.com/



对于coze和eino这种较新文档较少的开源库可以省去一部分代码阅读时间，直接定位具体问题