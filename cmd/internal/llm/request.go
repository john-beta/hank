package llm

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

// buildParams maps ToolResults to OpenAI's function_call_output input items
// re-fed against the previous response; otherwise Input is sent as plain text.
// The stored Prompt supplies instructions, tools, model, etc.
func buildParams(req Request) responses.ResponseNewParams {
	prompt := responses.ResponsePromptParam{ID: req.Prompt.ID, Version: param.Opt[string]{Value: req.Prompt.Version}}

	params := responses.ResponseNewParams{
		Prompt:            prompt,
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
