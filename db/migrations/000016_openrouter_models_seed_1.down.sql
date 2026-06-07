DELETE FROM prompts WHERE name = 'dsr-meso' and version = 1;

DELETE FROM annotators WHERE model_id IN (
    SELECT m.id FROM models m where m.slug IN (
        'openai/gpt-4o-mini', 'anthropic/claude-3.5-haiku',
        'google/gemini-2.5-flash-lite', 'deepseek/deepseek-chat-v3.1'
    )
);

DELETE FROM models WHERE slug IN (
    'openai/gpt-4o-mini', 'anthropic/claude-3.5-haiku',
    'google/gemini-2.5-flash-lite', 'deepseek/deepseek-chat-v3.1'
);