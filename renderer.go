package v8ssr

import (
	"io/ioutil"
	v8 "rogchap.com/v8go"
)


type Renderer struct {
	source string
	compiledScriptCache *v8.CompilerCachedData
	events chan renderEvent

	threads []*RenderThread
}

func NewRendererFromFile(filename string) (result *Renderer) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return NewRenderer(string(src))
}

func NewRenderer(source string) (result *Renderer) {
	result.events = make(chan renderEvent, 10)
	result.source = source

	iso := v8.NewIsolate()

	script, err := iso.CompileUnboundScript(result.source, "app.js", v8.CompileOptions{}) // compile script to get cached data
	if err != nil {
		panic(err)
	}

	result.compiledScriptCache = script.CreateCodeCache()

	result.threads = append(result.threads, result.newRenderThread())

	iso.Dispose()

	return
}

func (r *Renderer) Render(params interface{}) *renderResult {
	ret := make(chan *renderResult)
	defer close(ret)

	r.events <- renderEvent{
		action: request,
		params: params,
	}

	return <-ret
}

func (r *Renderer) Shutdown() {
	ret := make(chan *renderResult)
	defer close(ret)

	r.events <- renderEvent{
		action: shutdown,
	}

	<- ret
}