package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	defaultLangPath = "./lang"
	defaultLang     = "en-US"
	defaultEnvKey   = "RUN_MODE"
)

type Option func(*option)

type option struct {
	langDir     string
	defaultLang string
	envKey      string
	debugMode   bool
}

type Manager struct {
	LangList map[string]map[string]string
	Option   *option
	RunEnv   string
}

// result 输出结果
type result struct {
	Status int         `json:"status"`
	Desc   string      `json:"desc"`
	Trace  string      `json:"trace"`
	Data   interface{} `json:"data"`
}

type Data struct {
	Params []string
	Data   interface{}
}

func WithLangDir(dir string) Option {
	return func(o *option) {
		o.langDir = dir
	}
}

func WithDefaultLang(lang string) Option {
	return func(o *option) {
		o.defaultLang = lang
	}
}

func WithEnvKey(key string) Option {
	return func(o *option) {
		o.envKey = key
	}
}
func WithDebugMode(mode bool) Option {
	return func(o *option) {
		o.debugMode = mode
	}
}

func New(opts ...Option) (*Manager, error) {
	opt := &option{
		langDir:     defaultLangPath,
		defaultLang: defaultLang,
		envKey:      defaultEnvKey,
	}

	for _, f := range opts {
		f(opt)
	}

	langFileList, err := ioutil.ReadDir(opt.langDir)
	if err != nil {
		return nil, err
	}

	if len(langFileList) < 1 {
		return nil, errors.New("Initialize i18n config failed, Can't found lang config file! ")
	}

	langList := make(map[string]map[string]string)
	for i := range langFileList {
		langConfig := make(map[string]string)
		lang := strings.Split(langFileList[i].Name(), ".")[0]
		filePath := opt.langDir + "/" + langFileList[i].Name()
		langFile, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		byteValue, _ := ioutil.ReadAll(langFile)
		err = json.Unmarshal(byteValue, &langConfig)
		if err != nil {
			return nil, err
		}

		langList[lang] = langConfig
		_ = langFile.Close()
	}

	runEnv := os.Getenv(opt.envKey)

	return &Manager{LangList: langList, Option: opt, RunEnv: runEnv}, nil
}

// lang 返回语言
func (m *Manager) lang(c *gin.Context) string {
	lang := m.Option.defaultLang

	// 优先从header直接获取语言信息
	headerLang := c.GetHeader("lang")
	if headerLang != "" {
		return headerLang
	}

	ua := c.Request.UserAgent()
	for _, param := range strings.Split(ua, ";") {
		if !strings.Contains(param, "=") {
			continue
		}

		paramList := strings.Split(param, "=")
		switch paramList[0] {
		case "lang", "Lang":
			lang = paramList[1]
		}
	}

	return lang
}

// result 返回要输出的结果
func (m *Manager) result(c *gin.Context, status int, data interface{}, err error) result {
	var res result
	var tmplPrams []string

	res.Status = status

	switch data.(type) {
	case Data:
		d := data.(Data)
		res.Data = d.Data
		tmplPrams = d.Params
	default:
		res.Data = data
	}

	res.Desc = m.Trans(m.lang(c), strconv.Itoa(status), tmplPrams...)

	if m.isDebugMode(c) && err != nil {
		res.Trace = fmt.Sprintf("%v", err)
	}

	return res
}

// isDebugMode 返回该次请求结果是否支持Debug信息回传
// 非生产环境可以支持debug模式
// 优先级：运行环境>系统配置>请求配置
func (m *Manager) isDebugMode(c *gin.Context) bool {
	if m.RunEnv == "prod" {
		return false
	}

	if m.Option.debugMode {
		return true
	}

	// 从header中获取是否开启debug模式，非空则本次请求开启了debug模式
	debug := c.Request.Header.Get("debug")
	if debug != "" {
		return true
	}

	return false
}

// SetLang 修改默认语言
func (m *Manager) SetLang(lang string) {
	if lang != "" {
		m.Option.defaultLang = lang
	}
}

// Trans 输出翻译后的结果
func (m *Manager) Trans(lang string, code string, params ...string) string {
	l, ok := m.LangList[lang]
	// 如果取不到对应的语言列表，则使用默认语言列表
	if !ok {
		l = m.LangList[m.Option.defaultLang]
	}

	msg, ok := l[code]
	if ok {
		// 如果国际化模板有入参，则将入参渲染到模板里面
		if len(params) > 0 {
			var ps []interface{}
			for _, p := range params {
				ps = append(ps, p)
			}

			msg = fmt.Sprintf(msg, ps...)
		}

		return msg
	}

	return code
}

// Count 返回支持的国际化语言数量
func (m *Manager) Count() int {
	return len(m.LangList)
}

// Lang 返回支持的国际化语言
func (m *Manager) Lang() []string {
	var list []string
	for lang := range m.LangList {
		if lang != "" {
			list = append(list, lang)
		}
	}

	return list
}

// LangExist 返回指定语言支持结果
func (m *Manager) LangExist(lang string) bool {
	_, ok := m.LangList[lang]
	return ok
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (m *Manager) JSON(c *gin.Context, status int, data interface{}, err error) {
	c.Set("response_status", status)  //设置响应信息的statusCode
	c.JSON(http.StatusOK, m.result(c, status, data, err))
}

// JSONP serializes the given struct as JSON into the response body.
// It add padding to response body to request data from a server residing in a different domain than the client.
// It also sets the Content-Type as "application/javascript".
func (m *Manager) JSONP(c *gin.Context, status int, data interface{}, err error) {
	c.JSONP(http.StatusOK, m.result(c, status, data, err))
}

// AsciiJSON serializes the given struct as JSON into the response body with unicode to ASCII string.
// It also sets the Content-Type as "application/json".
func (m *Manager) AsciiJSON(c *gin.Context, status int, data interface{}, err error) {
	c.AsciiJSON(http.StatusOK, m.result(c, status, data, err))
}

// PureJSON serializes the given struct as JSON into the response body.
// PureJSON, unlike JSON, does not replace special html characters with their unicode entities.
func (m *Manager) PureJSON(c *gin.Context, status int, data interface{}, err error) {
	c.PureJSON(http.StatusOK, m.result(c, status, data, err))
}

// XML serializes the given struct as XML into the response body.
// It also sets the Content-Type as "application/xml".
func (m *Manager) XML(c *gin.Context, status int, data interface{}, err error) {
	c.XML(http.StatusOK, m.result(c, status, data, err))
}

// YAML serializes the given struct as YAML into the response body.
func (m *Manager) YAML(c *gin.Context, status int, data interface{}, err error) {
	c.YAML(http.StatusOK, m.result(c, status, data, err))
}
