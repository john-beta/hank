package llm

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// buildParams maps ToolResults to OpenAI's function_call_output input items
// re-fed against the previous response; otherwise Input is sent as plain text.
func buildParams(req Request) responses.ResponseNewParams {
	params := responses.ResponseNewParams{
		Model:             openai.ChatModelGPT4o,
		Instructions:      openai.String(req.Instructions),
		Tools:             mapTools(req.Tools),
		ParallelToolCalls: openai.Bool(false),
	}

	if req.PrevResponseID != "" {
		params.PreviousResponseID = openai.String(req.PrevResponseID)
	}

	if len(req.ToolResults) > 0 {
		items := make([]responses.ResponseInputItemUnionParam, 0, len(req.ToolResults))
		for _, tr := range req.ToolResults {
			items = append(items, responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: tr.CallID,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
						OfString: openai.String(tr.Output),
					},
				},
			})
		}
		params.Input = responses.ResponseNewParamsInputUnion{OfInputItemList: items}
	} else {
		params.Input = responses.ResponseNewParamsInputUnion{OfString: openai.String(req.Input)}
	}

	return params
}

func mapTools(defs []ToolDef) []responses.ToolUnionParam {
	if len(defs) == 0 {
		return nil
	}
	tools := make([]responses.ToolUnionParam, 0, len(defs))
	for _, d := range defs {
		// AutoReFeed is intentionally not copied: it's agent metadata, not part
		// of the OpenAI tool definition.
		tools = append(tools, responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name:        d.Name,
				Description: openai.String(d.Description),
				Parameters:  d.Parameters,
			},
		})
	}
	return tools
}
