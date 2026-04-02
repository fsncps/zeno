# TODOS

## P2 - Multi-provider AI Support

- **What:** Add support for Anthropic Claude, local LLMs (Ollama), and custom OpenAI-compatible endpoints.
- **Why:** Users want choice of AI providers for description generation, not just OpenAI.
- **Pros:** Broader user base, privacy options (local models), cost options.
- **Cons:** More code, more API integrations to maintain.
- **Context:** Currently AI is hardcoded to OpenAI GPT-4o-mini. Config has `OPENAI_API_KEY` in .env. Need provider selection in config.yaml and provider-specific clients in internal/ai/.
- **Effort:** M
- **Priority:** P2
- **Depends on:** Config system (completed)