package v8ssr

import (
	"log"
	v8 "rogchap.com/v8go"
)

type action int

const (
	request  action = iota
	shutdown action = iota
)

type renderResult struct {
	Output string
	Error error
}

type renderEvent struct {
	action
	result chan *renderResult
	params interface{}
}

type RenderThread struct {
	events chan renderEvent
	script *v8.UnboundScript
	isolate *v8.Isolate
}

func (r *Renderer) newRenderThread() *RenderThread {
	iso := v8.NewIsolate() // create a new JavaScript VM

	script, err := iso.CompileUnboundScript(r.source, "app.js", v8.CompileOptions{CachedData: r.compiledScriptCache}) // compile script in new isolate with cached data
	if err != nil {
		panic(err)
	}

	thread := &RenderThread{
		events: r.events,
		script: script,
		isolate: iso,
	}

	go thread.run()

	return thread
}

func (t *RenderThread) run() {
	for true {
		select {
		case event := <-t.events:
			switch event.action {
			case request:
				log.Printf("Render request: %v", event.params)
				break
			case shutdown:
				t.isolate.Dispose()
				return
			}
		}
	}
}