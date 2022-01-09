package v8ssr

import (
	"io/ioutil"
	v8 "rogchap.com/v8go"
)

type RendererConfig struct {
	Entry string `json:"entry"`
	Threads int `json:"threads"`
}

var DefaultRendererConfig RendererConfig = RendererConfig{
	Entry: "render()",
	Threads: 4,
}

type Renderer struct {
	Config RendererConfig

	source string
	compiledScriptCache *v8.CompilerCachedData
	events chan renderEvent
	threads []*RenderThread
}

func NewRendererFromFile(filename string, config RendererConfig) (result *Renderer) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return NewRenderer(string(src), config)
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

	result.events = make(chan renderEvent, 10)
	result.source = source

	iso := v8.NewIsolate()

	script, err := iso.CompileUnboundScript(result.source, "app.js", v8.CompileOptions{}) // compile script to get cached data
	if err != nil {
		panic(err)
	}

	result.compiledScriptCache = script.CreateCodeCache()

	for i := 0; i < result.Config.Threads; i++ {
		result.threads = append(result.threads, result.newRenderThread())
	}

	iso.Dispose()

	return
}

func (r *Renderer) Render(params interface{}) *renderResult {
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

	close(r.events)
}