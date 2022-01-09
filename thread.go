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
	renderer *Renderer
	events chan renderEvent
	isolate *v8.Isolate
	context *v8.Context
	script *v8.UnboundScript
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
		renderer: r,
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
				var result renderResult

				ctx := v8.NewContext(t.isolate)
				_, err := t.script.Run(ctx)
				if err != nil {
					result.Error = err
				} else {
					global := ctx.Global()
					err = global.Set("params", event.params)
					if err != nil {
						result.Error = err
					} else {
						value, err := ctx.RunScript(t.renderer.Config.Entry, "app.js")

						if err != nil {
							result.Error = err
						}

						if value != nil {
							result.Output = value.String()
						}
					}
				}

				event.result <- &result
				ctx.Close()

				break
			case shutdown:
				t.isolate.Dispose()

				event.result <- &renderResult{}
				return
			}
		}
	}
}