package v8ssr

import (
	"io/ioutil"
	"os"
	v8 "rogchap.com/v8go"
	"time"
)

type RendererConfig struct {
	filename string
	ReloadOnChange bool
	Entry string
	Threads int

}

var DefaultRendererConfig RendererConfig = RendererConfig{
	Entry: "entry()",
	filename: "",
	ReloadOnChange: false,
	Threads: 4,
}

type Renderer struct {
	Config RendererConfig

	source string
	compiledScriptCache *v8.CompilerCachedData
	events chan renderEvent
	threads []*RenderThread

	fileSize int64
	modTime time.Time
}

func NewRendererFromFile(filename string, config RendererConfig) (result *Renderer) {
	src := loadScriptFromFile(filename)
	config.filename = filename

	return NewRenderer(string(src), config)
}

func loadScriptFromFile(filename string) string {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(src)
}

func NewRenderer(source string, config RendererConfig) (result *Renderer) {
	result = &Renderer{}

	result.Config = DefaultRendererConfig

	if config.Entry != "" {
		result.Config.Entry = config.Entry
	}

	if config.Threads > 0 {
		result.Config.Threads = config.Threads
	}

	if config.ReloadOnChange && config.filename != "" {
		result.Config.ReloadOnChange = true
		result.Config.filename = config.filename
	} else if config.ReloadOnChange {
		panic("Must use NewRendererFromFile with ReloadOnChange")
	}

	result.source = source
	result.initializeThreads()

	return
}

func (r *Renderer) Render(params interface{}) *renderResult {
	if r.Config.ReloadOnChange {
		r.reloadIfChanged()
	}

	ret := make(chan *renderResult)
	defer close(ret)

	r.events <- renderEvent{
		result: ret,
		action: request,
		params: params,
	}

	return <-ret
}

func (r *Renderer) Shutdown() {
	for i := 0; i < r.Config.Threads; i++ {
		ret := make(chan *renderResult)

		r.events <- renderEvent{
			result: ret,
			action: shutdown,
		}

		<- ret
		close(ret)
	}

	r.threads = []*RenderThread{}

	close(r.events)
}

func (r *Renderer) initializeThreads() {
	r.events = make(chan renderEvent, 10)

	iso := v8.NewIsolate()

	script, err := iso.CompileUnboundScript(r.source, "app.js", v8.CompileOptions{}) // compile script to get cached data
	if err != nil {
		panic(err)
	}

	r.compiledScriptCache = script.CreateCodeCache()

	for i := 0; i < r.Config.Threads; i++ {
		r.threads = append(r.threads, r.newRenderThread())
	}

	iso.Dispose()
}

func (r *Renderer) reloadIfChanged() {
	stat, err := os.Stat(r.Config.filename)
	if err != nil {
		panic(err)
	}

	if stat.Size() != r.fileSize || stat.ModTime() != r.modTime {
		r.source = loadScriptFromFile(r.Config.filename)
		r.Shutdown()
		r.initializeThreads()
		r.modTime = stat.ModTime()
		r.fileSize = stat.Size()
	}
}