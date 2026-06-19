DELETE FROM annotators
WHERE model_id IN (SELECT id FROM models WHERE slug IN ('openai/gpt-4o', 'google/gemini-2.5-flash'));
 
DELETE FROM models WHERE slug IN ('openai/gpt-4o', 'google/gemini-2.5-flash');
 
DELETE FROM prompts WHERE name = 'image' AND version = 1;