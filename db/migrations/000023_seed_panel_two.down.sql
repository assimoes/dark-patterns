DELETE FROM annotators WHERE model_id IN (
    SELECT m.id FROM models m WHERE m.slug IN (
        'qwen/qwen3-next-80b-a3b-instruct:free', 'meta-llama/llama-3.3-70b-instruct:free',
        'deepseek/deepseek-v4-flash:free', 'openai/gpt-oss-120b:free'
    )
);

DELETE FROM models WHERE slug IN (
    'qwen/qwen3-next-80b-a3b-instruct:free', 'meta-llama/llama-3.3-70b-instruct:free',
    'deepseek/deepseek-v4-flash:free', 'openai/gpt-oss-120b:free'
);
