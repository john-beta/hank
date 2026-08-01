package llm

import "context"

// Stream runs a streaming Responses request as agent-level events. The channel
// closes when the stream ends, ctx is cancelled, or an error occurs; the reader
// goroutine is bound to ctx so it never outlives its consumer.
func (c *OpenAIClient) Stream(ctx context.Context, req Request) (<-chan StreamEvent, error) {
	stream := c.client.Responses.NewStreaming(ctx, buildParams(req))
	events := make(chan StreamEvent)

	go func() {
		defer close(events)

		// Function calls assembled across three stream phases, keyed by item id:
		// output_item.added (name + call id) -> arguments.delta -> arguments.done.
		pending := make(map[string]*FunctionCallData)

		for stream.Next() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			event := stream.Current()
			switch event.Type {
			case "response.output_text.delta":
				delta := event.AsResponseOutputTextDelta()
				c.send(ctx, events, StreamEvent{Type: "text_delta", Text: delta.Delta})

			case "response.output_item.added":
				item := event.AsResponseOutputItemAdded().Item

				switch item.Type {

				case "function_call":
					pending[item.ID] = &FunctionCallData{Name: item.Name, CallID: item.CallID}

				case "reasoning":
					c.send(ctx, events, StreamEvent{Type: "reasoning_start"})
				}

			case "response.output_item.done":
				if event.AsResponseOutputItemDone().Item.Type == "reasoning" {
					c.send(ctx, events, StreamEvent{Type: "reasoning_done"})
				}

			case "response.reasoning_summary_text.delta":
				delta := event.AsResponseReasoningSummaryTextDelta()
				c.send(ctx, events, StreamEvent{Type: "reasoning_delta", Text: delta.Delta})

			case "response.function_call_arguments.delta":
				d := event.AsResponseFunctionCallArgumentsDelta()
				if fc := pending[d.ItemID]; fc != nil {
					fc.Arguments += d.Delta
				}

			case "response.function_call_arguments.done":
				done := event.AsResponseFunctionCallArgumentsDone()
				fc := pending[done.ItemID]
				if fc == nil {
					continue // unknown item id; nothing to emit
				}
				fc.Arguments = done.Arguments
				c.send(ctx, events, StreamEvent{Type: "function_call", FunctionCall: fc})
				delete(pending, done.ItemID)

			case "response.completed":
				completed := event.AsResponseCompleted()
				c.send(ctx, events, StreamEvent{Type: "done", ResponseID: completed.Response.ID})
			}
		}

		if err := stream.Err(); err != nil {
			c.send(ctx, events, StreamEvent{Type: "error", Text: err.Error()})
		}
	}()

	return events, nil
}

// send delivers an event unless ctx is cancelled first, so the goroutine never
// blocks on an abandoned channel.
func (c *OpenAIClient) send(ctx context.Context, ch chan<- StreamEvent, ev StreamEvent) {
	select {
	case ch <- ev:
	case <-ctx.Done():
	}
}
