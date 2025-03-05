// Copyright 2024 Seakee. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package i18n provides internationalization support for Go applications.
// It allows for easy translation of messages based on language settings,
// and integrates with the Gin web framework for HTTP response handling.
// The package supports loading language files from a directory, and provides
// various response formats including JSON, XML, YAML, etc.
package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// defaultLangPath is the default directory path for language files
	defaultLangPath = "./lang"
	// defaultLang is the default language code used when no language is specified
	defaultLang = "en-US"
	// defaultEnvKey is the environment variable key used to determine the running mode
	defaultEnvKey = "RUN_MODE"
)

type (
	// Option is a function type that modifies the option struct
	Option func(*option)

	// option contains configuration settings for the i18n manager
	option struct {
		langDir     string // Directory path for language files
		defaultLang string // Default language code
		envKey      string // Environment variable key for run mode
		debugMode   bool   // Whether debug mode is enabled
	}

	// Manager handles internationalization operations and language file management
	Manager struct {
		LangList map[string]map[string]string // Map of language codes to their message maps
		Option   *option                      // Configuration options
		RunEnv   string                       // Current running environment
	}

	// result represents the standardized API response structure
	result struct {
		Code  int         `json:"code"`  // Response code
		Msg   string      `json:"msg"`   // Response message (translated)
		Trace trace       `json:"trace"` // Trace information for debugging
		Data  interface{} `json:"data"`  // Response data payload
	}

	// trace contains debugging information for API responses
	trace struct {
		ID   string `json:"id"`   // Trace identifier
		Desc string `json:"desc"` // Error description (only in debug mode)
	}

	// Data is a wrapper for response data that includes template parameters
	Data struct {
		Params []string    // Parameters for message template
		Data   interface{} // Actual response data
	}
)

// WithLangDir returns an Option that sets the language directory path.
//
// Parameters:
//   - dir: The directory path where language files are stored
//
// Returns:
//   - Option: A function that sets the language directory in the options
//
// Example:
//
//	i18n.New(i18n.WithLangDir("./custom/lang/path"))
func WithLangDir(dir string) Option {
	return func(o *option) {
		o.langDir = dir
	}
}

// WithDefaultLang returns an Option that sets the default language.
//
// Parameters:
//   - lang: The language code to use as default (e.g., "en-US", "zh-CN")
//
// Returns:
//   - Option: A function that sets the default language in the options
//
// Example:
//
//	i18n.New(i18n.WithDefaultLang("zh-CN"))
func WithDefaultLang(lang string) Option {
	return func(o *option) {
		o.defaultLang = lang
	}
}

// WithEnvKey returns an Option that sets the environment variable key.
//
// Parameters:
//   - key: The environment variable key used to determine the running mode
//
// Returns:
//   - Option: A function that sets the environment key in the options
//
// Example:
//
//	i18n.New(i18n.WithEnvKey("APP_ENV"))
func WithEnvKey(key string) Option {
	return func(o *option) {
		o.envKey = key
	}
}

// WithDebugMode returns an Option that sets the debug mode.
//
// Parameters:
//   - mode: Boolean value indicating whether debug mode is enabled
//
// Returns:
//   - Option: A function that sets the debug mode in the options
//
// Example:
//
//	i18n.New(i18n.WithDebugMode(true))
func WithDebugMode(mode bool) Option {
	return func(o *option) {
		o.debugMode = mode
	}
}

// New initializes and returns a new Manager instance with the provided options.
// It loads language files from the configured directory and sets up the Manager
// with the specified options.
//
// Parameters:
//   - opts: Variable number of Option functions to configure the Manager
//
// Returns:
//   - *Manager: A pointer to the initialized Manager instance
//   - error: An error if initialization fails, nil otherwise
//
// Example:
//
//	manager, err := i18n.New(
//	    i18n.WithLangDir("./lang"),
//	    i18n.WithDefaultLang("en-US"),
//	    i18n.WithDebugMode(true),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(opts ...Option) (*Manager, error) {
	// Initialize options with default values
	opt := &option{
		langDir:     defaultLangPath,
		defaultLang: defaultLang,
		envKey:      defaultEnvKey,
	}

	// Apply all provided option functions
	for _, f := range opts {
		f(opt)
	}

	// Load language files from the specified directory
	langList, err := loadLangFiles(opt.langDir)
	if err != nil {
		return nil, err
	}

	// Ensure at least one language file was loaded
	if len(langList) == 0 {
		return nil, errors.New("没有找到语言配置文件")
	}

	// Get the current running environment from environment variables
	runEnv := os.Getenv(opt.envKey)

	// Create and return the Manager instance
	return &Manager{LangList: langList, Option: opt, RunEnv: runEnv}, nil
}

// loadLangFiles reads and parses language files from the specified directory.
// It walks through the directory, reads each file, and parses its JSON content
// into a map of message keys to translated messages.
//
// Parameters:
//   - langDir: The directory path where language files are stored
//
// Returns:
//   - map[string]map[string]string: A map of language codes to their message maps
//   - error: An error if reading or parsing files fails, nil otherwise
func loadLangFiles(langDir string) (map[string]map[string]string, error) {
	// Initialize the map to store language configurations
	langList := make(map[string]map[string]string)

	// Walk through the language directory
	err := filepath.Walk(langDir, func(path string, info os.FileInfo, err error) error {
		// Skip directories, only process files
		if !info.IsDir() {
			// Extract language code from filename (without extension)
			lang := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			langConfig := make(map[string]string)

			// Read file content
			var byteValue []byte
			byteValue, err = os.ReadFile(path)
			if err != nil {
				return err
			}

			// Parse JSON content into the language configuration map
			if err = json.Unmarshal(byteValue, &langConfig); err != nil {
				return err
			}

			// Store the language configuration in the language list
			langList[lang] = langConfig
		}
		return nil
	})

	return langList, err
}

// lang determines the language to use for the current request.
// It first checks the "lang" header, then looks for a "lang" parameter
// in the User-Agent string, and falls back to the default language.
//
// Parameters:
//   - c: The Gin context containing request information
//
// Returns:
//   - string: The language code to use for the current request
func (m *Manager) lang(c *gin.Context) string {
	// First priority: Check for "lang" header
	headerLang := c.Request.Header.Get("lang")
	if headerLang != "" {
		return headerLang
	}

	// Second priority: Check for "lang" parameter in User-Agent
	ua := c.Request.UserAgent()
	for _, param := range strings.Split(ua, ";") {
		paramList := strings.Split(param, "=")
		if len(paramList) == 2 && paramList[0] == "lang" {
			return paramList[1]
		}
	}

	// Fallback to default language
	return m.Option.defaultLang
}

// result creates a standardized response structure for API responses.
// It handles translation of messages, inclusion of trace information,
// and formatting of response data.
//
// Parameters:
//   - c: The Gin context containing request information
//   - code: The response code (used for message lookup)
//   - data: The response data or Data struct with template parameters
//   - err: An error to include in trace information (if debug mode is enabled)
//
// Returns:
//   - result: The formatted response structure
func (m *Manager) result(c *gin.Context, code int, data interface{}, err error) result {
	var res result
	var tmplPrams []string

	// Set response code
	res.Code = code

	// Handle data and extract template parameters if provided
	switch d := data.(type) {
	case Data:
		// If data is a Data struct, extract template parameters and actual data
		res.Data = d.Data
		tmplPrams = d.Params
	default:
		// Otherwise, use data as-is
		res.Data = data
	}

	// Translate the message using the determined language and code
	res.Msg = m.Trans(m.lang(c), strconv.Itoa(code), tmplPrams...)

	// Include trace ID if available in the context
	traceID, exists := c.Get("trace_id")
	if exists {
		res.Trace.ID = traceID.(string)
	}

	// Include error description in trace if debug mode is enabled
	if m.isDebugMode(c) && err != nil {
		res.Trace.Desc = fmt.Sprintf("%v", err)
	}

	return res
}

// isDebugMode determines whether debug information should be included in responses.
// The decision is based on a priority hierarchy:
// 1. Production environment always disables debug mode
// 2. If not in production, check if debug mode is enabled in options
// 3. If not enabled in options, check for "debug" header in the request
//
// Parameters:
//   - c: The Gin context containing request information
//
// Returns:
//   - bool: true if debug mode is enabled, false otherwise
func (m *Manager) isDebugMode(c *gin.Context) bool {
	// Production environment always disables debug mode
	if m.RunEnv == "prod" {
		return false
	}

	// If debug mode is enabled in options, enable it
	if m.Option.debugMode {
		return true
	}

	// Check for "debug" header in the request
	debug := c.Request.Header.Get("debug")

	return debug != ""
}

// SetLang changes the default language for the Manager.
// If an empty string is provided, the default language remains unchanged.
//
// Parameters:
//   - lang: The new default language code
//
// Example:
//
//	manager.SetLang("fr-FR")
func (m *Manager) SetLang(lang string) {
	if lang != "" {
		m.Option.defaultLang = lang
	}
}

// Trans translates a message code to a localized message in the specified language.
// If the language is not supported, it falls back to the default language.
// If template parameters are provided, they are formatted into the message.
//
// Parameters:
//   - lang: The language code to use for translation
//   - code: The message code to translate
//   - params: Optional parameters to format into the message template
//
// Returns:
//   - string: The translated message, or the original code if no translation is found
//
// Example:
//
//	// Assuming "1001" maps to "Hello, %s!" in the language files
//	message := manager.Trans("en-US", "1001", "World")
//	// message will be "Hello, World!"
func (m *Manager) Trans(lang string, code string, params ...string) string {
	// Get the message map for the specified language
	l, ok := m.LangList[lang]
	// If the language is not supported, fall back to the default language
	if !ok {
		l = m.LangList[m.Option.defaultLang]
	}

	// Look up the message for the specified code
	msg, ok := l[code]
	if ok {
		// If template parameters are provided, format them into the message
		if len(params) > 0 {
			var ps []interface{}
			// Convert string parameters to interface{} for fmt.Sprintf
			for _, p := range params {
				ps = append(ps, p)
			}

			// Format the message with the parameters
			msg = fmt.Sprintf(msg, ps...)
		}

		return msg
	}

	// If no translation is found, return the original code
	return code
}

// Count returns the number of supported languages.
//
// Returns:
//   - int: The number of languages in the language list
//
// Example:
//
//	count := manager.Count()
//	fmt.Printf("Supported languages: %d\n", count)
func (m *Manager) Count() int {
	return len(m.LangList)
}

// Lang returns a list of all supported language codes.
//
// Returns:
//   - []string: A slice containing all supported language codes
//
// Example:
//
//	languages := manager.Lang()
//	fmt.Printf("Supported languages: %v\n", languages)
func (m *Manager) Lang() []string {
	var list []string
	// Iterate through all language codes in the language list
	for lang := range m.LangList {
		if lang != "" {
			list = append(list, lang)
		}
	}

	return list
}

// LangExist checks if a specific language is supported.
//
// Parameters:
//   - lang: The language code to check
//
// Returns:
//   - bool: true if the language is supported, false otherwise
//
// Example:
//
//	if manager.LangExist("fr-FR") {
//	    fmt.Println("French is supported")
//	}
func (m *Manager) LangExist(lang string) bool {
	_, ok := m.LangList[lang]
	return ok
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
//
// Parameters:
//   - c: The Gin context for the HTTP response
//   - code: The response code (used for message lookup)
//   - data: The response data or Data struct with template parameters
//   - err: An error to include in trace information (if debug mode is enabled)
//
// Example:
//
//	func HandleRequest(c *gin.Context) {
//	    data := someData()
//	    manager.JSON(c, 200, data, nil)
//	}
func (m *Manager) JSON(c *gin.Context, code int, data interface{}, err error) {
	// Store the response code in the context for potential middleware use
	c.Set("response_code", code)
	// Send JSON response with standardized structure
	c.JSON(http.StatusOK, m.result(c, code, data, err))
}

// JSONP serializes the given struct as JSON into the response body with JSONP support.
// It enables cross-origin requests by wrapping the JSON response in a JavaScript callback function.
//
// Parameters:
//   - c: The Gin context for the HTTP response
//   - code: The response code (used for message lookup)
//   - data: The response data or Data struct with template parameters
//   - err: An error to include in trace information (if debug mode is enabled)
//
// Example:
//
//	func HandleJSONPRequest(c *gin.Context) {
//	    data := someData()
//	    manager.JSONP(c, 200, data, nil)
//	    // Response: callback({"code": 200, "msg": "...", "data": ...})
//	}
func (m *Manager) JSONP(c *gin.Context, code int, data interface{}, err error) {
	c.JSONP(http.StatusOK, m.result(c, code, data, err))
}

// AsciiJSON serializes the given struct as JSON into the response body,
// converting any non-ASCII characters to their Unicode escape sequences.
//
// Parameters:
//   - c: The Gin context for the HTTP response
//   - code: The response code (used for message lookup)
//   - data: The response data or Data struct with template parameters
//   - err: An error to include in trace information (if debug mode is enabled)
//
// Example:
//
//	func HandleAsciiJSONRequest(c *gin.Context) {
//	    data := map[string]string{"name": "世界"}
//	    manager.AsciiJSON(c, 200, data, nil)
//	    // Response: {"name":"\u4e16\u754c"}
//	}
func (m *Manager) AsciiJSON(c *gin.Context, code int, data interface{}, err error) {
	c.AsciiJSON(http.StatusOK, m.result(c, code, data, err))
}

// PureJSON serializes the given struct as JSON into the response body,
// preserving special HTML characters without Unicode escaping.
//
// Parameters:
//   - c: The Gin context for the HTTP response
//   - code: The response code (used for message lookup)
//   - data: The response data or Data struct with template parameters
//   - err: An error to include in trace information (if debug mode is enabled)
//
// Example:
//
//	func HandlePureJSONRequest(c *gin.Context) {
//	    data := map[string]string{"html": "<p>Hello</p>"}
//	    manager.PureJSON(c, 200, data, nil)
//	    // Response: {"html":"<p>Hello</p>"}
//	}
func (m *Manager) PureJSON(c *gin.Context, code int, data interface{}, err error) {
	c.PureJSON(http.StatusOK, m.result(c, code, data, err))
}

// XML serializes the given struct as XML into the response body.
// The response will include an XML header and proper content type.
//
// Parameters:
//   - c: The Gin context for the HTTP response
//   - code: The response code (used for message lookup)
//   - data: The response data or Data struct with template parameters
//   - err: An error to include in trace information (if debug mode is enabled)
//
// Example:
//
//	func HandleXMLRequest(c *gin.Context) {
//	    data := struct {
//	        Name string `xml:"name"`
//	    }{Name: "test"}
//	    manager.XML(c, 200, data, nil)
//	    // Response: <?xml version="1.0" encoding="UTF-8"?>
//	    //          <response>
//	    //              <code>200</code>
//	    //              <msg>Success</msg>
//	    //              <data>
//	    //                  <name>test</name>
//	    //              </data>
//	    //          </response>
//	}
func (m *Manager) XML(c *gin.Context, code int, data interface{}, err error) {
	c.XML(http.StatusOK, m.result(c, code, data, err))
}

// YAML serializes the given struct as YAML into the response body.
// The response will use proper YAML formatting and indentation.
//
// Parameters:
//   - c: The Gin context for the HTTP response
//   - code: The response code (used for message lookup)
//   - data: The response data or Data struct with template parameters
//   - err: An error to include in trace information (if debug mode is enabled)
//
// Example:
//
//	func HandleYAMLRequest(c *gin.Context) {
//	    data := map[string]string{"name": "test"}
//	    manager.YAML(c, 200, data, nil)
//	    // Response:
//	    // code: 200
//	    // msg: Success
//	    // data:
//	    //   name: test
//	}
func (m *Manager) YAML(c *gin.Context, code int, data interface{}, err error) {
	c.YAML(http.StatusOK, m.result(c, code, data, err))
}
