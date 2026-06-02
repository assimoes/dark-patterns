INSERT INTO taxonomy_high_levels (name, description, version) VALUES
('Temporal Manipulation', 'patterns that exploit time and pacing', 1),
('Predatory Monetization', 'patterns that push or obscure spending', 1),
('Social Exploitation', 'patterns that weaponise social ties and competition', 1),
('Psychological Exploitation', 'patterns that exploit cognitive and emotional biases', 1),
('Deceptive Representation', 'patterns that misrepresent the product', 1)
ON CONFLICT DO NOTHING;

