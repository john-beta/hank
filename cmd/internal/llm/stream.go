package llm

import "context"

// Stream opens a streaming Responses request and returns a channel of
// agent-level events. The returned channel is closed when the stream ends,
// the context is cancelled, or an error occurs. The reader goroutine is bound
// to ctx so it never outlives its consumer.
func (c *OpenAIClient) Stream(ctx context.Context, req Request) (<-chan StreamEvent, error) {
	stream := c.client.Responses.NewStreaming(ctx, buildParams(req))
	events := make(chan StreamEvent)

	go func() {
		defer close(events)

		// In-progress function call, assembled across three stream phases:
		// output_item.added (name + call id) -> arguments.delta (chunks) ->
		// arguments.done (final arguments, emit).
		var fcName, fcCallID, fcArgs string

		for stream.Next() {
			// Bail out promptly if the consumer disconnected.
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
				if item.Type == "function_call" {
					fcName = item.Name
					fcCallID = item.CallID
					fcArgs = ""
				}

			case "response.function_call_arguments.delta":
				fcArgs += event.AsResponseFunctionCallArgumentsDelta().Delta

			case "response.function_call_arguments.done":
				fcArgs = event.AsResponseFunctionCallArgumentsDone().Arguments
				c.send(ctx, events, StreamEvent{
					Type: "function_call",
					FunctionCall: &FunctionCallData{
						CallID:    fcCallID,
						Name:      fcName,
						Arguments: fcArgs,
					},
				})
				fcName, fcCallID, fcArgs = "", "", ""

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

// send delivers an event unless the context is cancelled first, so the reader
// goroutine can never block forever on an abandoned channel.
func (c *OpenAIClient) send(ctx context.Context, ch chan<- StreamEvent, ev StreamEvent) {
	select {
	case ch <- ev:
	case <-ctx.Done():
	}
}
