package v8ssr

import (
	"context"
	"fmt"
	"log"
	v8 "rogchap.com/v8go"
)

type action int

const (
	request  action = iota
	shutdown action = iota
)

var threadCount int

type renderResult struct {
	Output string
	Error error
}

type renderEvent struct {
	action
	context context.Context
	result chan *renderResult
	params interface{}
}

type RenderThread struct {
	id int
	renderer *Renderer
	events chan renderEvent
	isolate *v8.Isolate
	context context.Context
	script *v8.UnboundScript
	global *v8.ObjectTemplate
}

func (r *Renderer) newRenderThread() *RenderThread {
	iso := v8.NewIsolate() // create a new JavaScript VM

	script, err := iso.CompileUnboundScript(r.source, "app.js", v8.CompileOptions{CachedData: r.compiledScriptCache}) // compile script in new isolate with cached data
	if err != nil {
		panic(err)
	}

	global := v8.NewObjectTemplate(iso) // a template that represents a JS Object

	thread := &RenderThread{
		events: r.events,
		script: script,
		isolate: iso,
		renderer: r,
		global: global,
		id: threadCount,
	}

	threadCount += 1

	for name, f := range r.callbacks {
		fun := v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
			r := f(thread.context, info.Args())
			result, err := v8.NewValue(iso, r)
			if err != nil {
				panic(fmt.Errorf("callback %s returned value %v, cannot be converted to v8 - %v", name, r, err))
			}
			return result
		})
		global.Set(name, fun)
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
				log.Printf("ID %d, Render request: %v", t.id, event.params)
				var result renderResult

				t.context = event.context

				ctx := v8.NewContext(t.isolate, t.global)
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