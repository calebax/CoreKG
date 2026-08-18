package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "consumes": [
        "application/json"
    ],
    "produces": [
        "application/json"
    ],
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/forest.BatchGetFile": {
            "post": {
                "summary": "批量查询文档",
                "description": "根据文档 ID 列表批量获取文档信息"
            }
        },
        "/keapi.GetFileChunks": {
            "post": {
                "summary": "根据知识库文件 ID 和 chunk 序号列表查询 chunk 信息",
                "description": "根据知识库文件 ID 和 chunk 序号列表查询 chunk 信息"
            }
        },
        "/forest.CreateDir": {
            "post": {
                "summary": "创建文件夹",
                "description": "创建知识库目录"
            }
        },
        "/forest.CreateForest": {
            "post": {
                "summary": "创建知识库",
                "description": "创建知识库，支持标准知识库与 Excel 知识库"
            }
        },
        "/forest.DeleteFile": {
            "post": {
                "summary": "删除文档",
                "description": "删除知识库文档"
            }
        },
        "/forest.DeleteForest": {
            "post": {
                "summary": "删除知识库",
                "description": "删除知识库及其关联资源"
            }
        },
        "/forest.DeletePath": {
            "post": {
                "summary": "删除文件夹",
                "description": "删除目录及其子资源"
            }
        },
        "/forest.ListFile": {
            "post": {
                "summary": "列出文档列表",
                "description": "获取知识库下的文档列表"
            }
        },
        "/forest.ListForest": {
            "post": {
                "summary": "列出知识库列表",
                "description": "获取当前 API Key 所属主体可访问的知识库列表"
            }
        },
        "/forest.PreviewFileByURL": {
            "post": {
                "summary": "下载文档",
                "description": "获取文档预签名下载地址"
            }
        },
        "/forest.RenamePath": {
            "post": {
                "summary": "重命名文件夹",
                "description": "重命名知识库目录或文件"
            }
        },
        "/forest.UpdateForestWithPerm": {
            "post": {
                "summary": "更新知识库",
                "description": "更新知识库基础信息并保留现有权限配置"
            }
        },
        "/forest.UploadFile": {
            "post": {
                "summary": "上传文档",
                "description": "通过 multipart/form-data 上传文档"
            }
        },
        "/kesearch.ForestSearch": {
            "post": {
                "summary": "内容检索",
                "description": "知识库内容检索"
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "description": "Bearer API Key",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1.0",
	Host:             "127.0.0.1:8086",
	BasePath:         "/v3",
	Schemes:          []string{"http"},
	Title:            "keapi API",
	Description:      "external knowledge api service",
	InfoInstanceName: "keapi",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
