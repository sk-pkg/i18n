# sk-pkg/i18n Documentation

## Introduction

`sk-pkg/i18n` is an internationalization package for the Gin framework. It provides simple and easy-to-use APIs for handling multi-language support, allowing your application to return messages according to the user's language preferences.

## Installation

```bash
go get -u "github.com/sk-pkg/i18n"
```

## Quick Start

### 1. Define Language Packs

First, you need to define internationalization language packs for your project. Language packs are JSON format files, with one file per language, named with the language code (e.g., `zh-CN.json`, `en-US.json`).

For example:

`zh-CN.json`:
```json
{
  "-1": "系统繁忙",
  "0": "ok",
  "500": "fail",
  "400": "请求参数错误",
  "1000": "你好,%s!你的账号是:%s"
}
```

`en-US.json`:
```json
{
  "-1": "System is busy",
  "0": "ok",
  "500": "fail",
  "400": "Request parameter error",
  "1000": "Hello,%s!Your account is:%s"
}
```

### 2. Initialize i18n Instance

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/sk-pkg/i18n"
    "log"
)

func main() {
    r := gin.Default()
    
    // Initialize i18n instance
    msg, err := i18n.New(
        i18n.WithLangDir("./lang"),  // Specify language pack directory
        i18n.WithDebugMode(true)     // Enable debug mode
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Set default language
    msg.SetLang("zh-CN")
    
    // Register routes...
    
    r.Run(":8080")
}
```

### 3. Use in Route Handlers

```go
// Basic usage
r.GET("/ok", func(c *gin.Context) {
    msg.JSON(c, 0, "success", nil)
})

// Error handling
r.GET("/busy", func(c *gin.Context) {
    msg.JSON(c, -1, "busy", errors.New("busy... "))
})

// Internationalization with parameters
r.GET("/test", func(c *gin.Context) {
    msg.JSON(c, 1000, i18n.Data{
        Params: []string{"Seakee", "18888888888"},
        Data:   "test",
    }, nil)
})
```

### 4. Direct Text Translation

```go
// Translate text in specific languages
enText := msg.Trans("en-US", "1000", "Seakee", "18888888888")
zhText := msg.Trans("zh-CN", "1000", "Seakee", "18888888888")

fmt.Println(enText) // Output: Hello,Seakee!Your account is:18888888888
fmt.Println(zhText) // Output: 你好,Seakee!你的账号是:18888888888
```

## Configuration Options

When initializing an i18n instance, you can configure it with the following options:

### 1. Language Pack Directory

Specify the directory where language pack files are stored, this is required:

```go
i18n.WithLangDir("./lang")
```

### 2. Default Language

Set the default language, defaults to `zh-CN`:

```go
i18n.WithDefaultLang("zh-CN")
```

You can also change the default language during runtime with:

```go
msg.SetLang("en-US")
```

### 3. Environment Variable Key

Set the environment variable key used to get the running environment, defaults to `RUN_MODE`:

```go
i18n.WithEnvKey("RUN_MODE")
```

### 4. Debug Mode

Set whether to enable debug mode, which will show detailed error information in the response, defaults to `false`:

```go
i18n.WithDebugMode(true)
```

## Response Methods

The i18n package provides multiple response methods to return internationalized messages in different formats:

### 1. JSON

Output in `application/json` format:

```go
msg.JSON(c *gin.Context, code int, data interface{}, err error)
```

### 2. JSONP

Output in `application/javascript` format:

```go
msg.JSONP(c *gin.Context, code int, data interface{}, err error)
```

### 3. AsciiJSON

Output in `application/json` format, converting Unicode characters to ASCII:

```go
msg.AsciiJSON(c *gin.Context, code int, data interface{}, err error)
```

### 4. PureJSON

Output in `application/json` format, without replacing special HTML characters with Unicode entities:

```go
msg.PureJSON(c *gin.Context, code int, data interface{}, err error)
```

### 5. XML

Output in `application/xml` format:

```go
msg.XML(c *gin.Context, code int, data interface{}, err error)
```

### 6. YAML

Output in `yaml` format:

```go
msg.YAML(c *gin.Context, code int, data interface{}, err error)
```

## Other Useful Methods

### 1. Get Supported Languages List

```go
languages := msg.Lang()
```

### 2. Get Number of Supported Languages

```go
count := msg.Count()
```

### 3. Check if a Language is Supported

```go
exists := msg.LangExist("zh-CN")
```

## Language Detection Mechanism

The i18n package detects the user's language preference in the following priority:

1. The `lang` field in the request header
2. The `lang` parameter in the User-Agent
3. The default language

## Response Format

All response methods will return data in the following format:

```json
{
    "code": 0,           // Status code
    "msg": "ok",         // Internationalized message
    "trace": {           // Debug information
        "id": "",        // Trace ID
        "desc": ""       // Error description (only shown in debug mode)
    },
    "data": "success"    // Business data
}
```

## Complete Example

```go
package main

import (
    "errors"
    "github.com/gin-gonic/gin"
    "github.com/sk-pkg/i18n"
    "log"
)

func main() {
    r := gin.Default()
    msg, err := i18n.New(
        i18n.WithLangDir("./lang"),
        i18n.WithDefaultLang("zh-CN"),
        i18n.WithDebugMode(true)
    )
    if err != nil {
        log.Fatal(err)
    }

    r.GET("/busy", func(c *gin.Context) {
        msg.XML(c, -1, "busy", errors.New("busy... "))
    })

    r.GET("/ok", func(c *gin.Context) {
        msg.JSON(c, 0, "success", nil)
    })

    r.GET("/fail", func(c *gin.Context) {
        msg.JSONP(c, 500, "fail", nil)
    })

    r.GET("/params", func(c *gin.Context) {
        msg.YAML(c, 400, "params", nil)
    })

    r.GET("/test", func(c *gin.Context) {
        msg.JSON(c, 1000, i18n.Data{
            Params: []string{"Seakee", "18888888888"},
            Data:   "test",
        }, nil)
    })

    r.Run(":8080")
}
```

## Notes

1. Language pack files must be in valid JSON format
2. Language pack filenames must be language codes (e.g., `zh-CN.json`, `en-US.json`)
3. In production environments, it's recommended to disable debug mode to avoid leaking sensitive information
4. If a corresponding language or message code cannot be found, the message code itself will be returned as the message content